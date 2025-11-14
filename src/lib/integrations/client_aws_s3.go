package integrations

import (
	"bytes"
	"context"
	sterrors "errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stormkit-io/stormkit-io/src/lib/config"
	"github.com/stormkit-io/stormkit-io/src/lib/errors"
	"github.com/stormkit-io/stormkit-io/src/lib/utils/file"
)

type S3Args struct {
	BucketName string
	KeyPrefix  string
	ACL        s3types.ObjectCannedACL
}

func (a *AWSClient) getFile(args GetFileArgs) (*GetFileResult, error) {
	bucketName, keyPrefix := a.parseS3Location(args.Location)

	out, err := a.S3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &keyPrefix,
	})

	if err != nil {
		var nsk *s3types.NoSuchKey

		if sterrors.As(err, &nsk) {
			return nil, nil
		}

		return nil, errors.Wrap(err, errors.ErrorTypeExternal, "failed to get S3 object").
			WithContext("bucket", bucketName).
			WithContext("key", keyPrefix)
	}

	if out == nil {
		return nil, nil
	}

	buf := bytes.NewBuffer(nil)

	if _, err := io.Copy(buf, out.Body); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to read S3 object body").
			WithContext("bucket", bucketName).
			WithContext("key", keyPrefix)
	}

	content := buf.Bytes()
	contentType := ""
	contentLength := int64(0)

	if out.ContentType != nil {
		contentType = *out.ContentType
	} else {
		contentType = DetectContentType(keyPrefix, content)
	}

	if out.ContentLength != nil {
		contentLength = *out.ContentLength
	}

	return &GetFileResult{
		Content:     content,
		ContentType: contentType,
		Size:        contentLength,
	}, nil
}

// ZipDownloader downloads the zip file with the given bucket and keyprefix.
// If the folder has been previously created, it returns the path immediately.
// If not, †his method will create a temp folder, download the zip in there,
// unzip it and remove the zip file.
func (a *AWSClient) ZipDownloader(deploymentID, bucket, keyprefix string) (string, error) {
	// First, check if the folder exists
	folder := fmt.Sprintf("d-%s", deploymentID)
	path := filepath.Join(os.TempDir(), folder)

	stat, err := os.Stat(path)

	if err == nil && stat != nil {
		return path, nil
	}

	err = os.Mkdir(path, 0775)

	if err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeInternal, "failed to create directory for deployment").
			WithContext("deployment_id", deploymentID).
			WithContext("path", path)
	}

	zipPath := filepath.Join(path, "sk-client.zip")
	f, err := os.Create(zipPath)

	if err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeInternal, "failed to create zip file").
			WithContext("deployment_id", deploymentID).
			WithContext("zip_path", zipPath)
	}

	defer f.Close()

	n, err := a.downloader.Download(context.TODO(), f, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(keyprefix),
	})

	if err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeExternal, "failed to download zip from S3").
			WithContext("deployment_id", deploymentID).
			WithContext("bucket", bucket).
			WithContext("key", keyprefix)
	}

	if n == 0 {
		return "", errors.New(errors.ErrorTypeExternal, "did not download any file").
			WithContext("deployment_id", deploymentID).
			WithContext("bucket", bucket).
			WithContext("key", keyprefix)
	}

	unzipOpts := file.UnzipOpts{
		ZipFile:    zipPath,
		ExtractDir: path,
		LowerCase:  false,
	}

	if err := file.Unzip(unzipOpts); err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeInternal, "failed to unzip file").
			WithContext("deployment_id", deploymentID).
			WithContext("zip_path", zipPath).
			WithContext("extract_dir", path)
	}

	if err := os.Remove(zipPath); err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeInternal, "failed to remove zip file").
			WithContext("deployment_id", deploymentID).
			WithContext("zip_path", zipPath)
	}

	return path, nil
}

func (a *AWSClient) serveFromZip(args GetFileArgs) (*GetFileResult, error) {
	a.mux.Lock()
	defer a.mux.Unlock()

	if a.zipManager == nil {
		a.zipManager = NewZipManager(a.ZipDownloader)
	}

	return a.zipManager.GetFile(args)
}

// GetFile returns a file from the bucket.
func (a *AWSClient) GetFile(args GetFileArgs) (*GetFileResult, error) {
	if strings.HasSuffix(args.Location, "sk-client.zip") {
		return a.serveFromZip(args)
	}

	// This is required to make old-style locations work.
	args.Location = path.Join(args.Location, args.FileName)

	return a.getFile(args)
}

func (a *AWSClient) bucketName(args UploadArgs) string {
	bucketName := args.BucketName

	if args.BucketName == "" {
		bucketName = config.Get().AWS.StorageBucket
	}

	pckg := args.AppPackage

	if config.IsStormkitCloud() && pckg != config.PackageFree && pckg != "" {
		bucketName = fmt.Sprintf("%s-versioned", bucketName)
	}

	return bucketName
}

func (a *AWSClient) uploadZipToS3(pathToZip string, args UploadArgs) (UploadOverview, error) {
	keyPrefix := fmt.Sprintf("%d/%d", args.AppID, args.DeploymentID)
	bucketName := a.bucketName(args)
	result := UploadOverview{}

	content, err := os.ReadFile(pathToZip)

	if err != nil {
		return result, errors.Wrap(err, errors.ErrorTypeInternal, "failed to read zip file").
			WithContext("app_id", args.AppID).
			WithContext("deployment_id", args.DeploymentID).
			WithContext("path", pathToZip)
	}

	stat, err := os.Stat(pathToZip)

	if err != nil {
		return result, errors.Wrap(err, errors.ErrorTypeInternal, "failed to get zip file info").
			WithContext("app_id", args.AppID).
			WithContext("deployment_id", args.DeploymentID).
			WithContext("path", pathToZip)
	}

	zipName := path.Base(pathToZip)
	size := stat.Size()
	file := File{
		Content:      content,
		ContentType:  DetectContentType(args.ClientZip, content),
		RelativePath: zipName, // to the key prefix
		Size:         size,
	}

	err = a.UploadFile(file, S3Args{
		BucketName: bucketName,
		KeyPrefix:  keyPrefix,
		ACL:        s3types.ObjectCannedACLPrivate,
	})

	if err != nil {
		return result, errors.Wrap(err, errors.ErrorTypeExternal, "failed to upload zip to S3").
			WithContext("app_id", args.AppID).
			WithContext("deployment_id", args.DeploymentID).
			WithContext("bucket", bucketName).
			WithContext("key_prefix", keyPrefix).
			WithContext("file", zipName)
	}

	result.BytesUploaded = size
	result.FilesUploaded = 1
	result.Location = fmt.Sprintf("aws:%s/%s/%s", bucketName, keyPrefix, zipName)
	return result, err
}

// UploadFile uploads a single file to S3 destination.
func (a *AWSClient) UploadFile(file File, s3args any) error {
	opts := s3args.(S3Args)
	filePath := filepath.Join(opts.KeyPrefix, file.RelativePath)

	if opts.ACL == "" {
		opts.ACL = s3types.ObjectCannedACLPrivate
	}

	input := &s3.PutObjectInput{
		Bucket:               &opts.BucketName,
		Key:                  &filePath,
		ContentType:          &file.ContentType,
		ContentLength:        &file.Size,
		ServerSideEncryption: s3types.ServerSideEncryptionAes256,
		// This is a required to allow reading the file through our CDN,
		// but it may create problems with custom storages.
		// Therefore, we may want to make this variable.
		ACL: opts.ACL,
	}

	if file.Content != nil {
		input.Body = bytes.NewReader(file.Content)
	} else {
		input.Body = file.Pointer
	}

	_, err := a.uploader.Upload(context.Background(), input)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeExternal, "failed to upload file to S3").
			WithContext("bucket", opts.BucketName).
			WithContext("key", filePath).
			WithContext("size", file.Size).
			WithContext("content_type", file.ContentType)
	}
	return nil
}

func (a *AWSClient) deleteS3Folder(ctx context.Context, bucketName, keyPrefix string) error {
	// List all objects in the folder
	listObjectsResp, err := a.S3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(keyPrefix), // Folder prefix (e.g., "my-folder/")
	})

	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeExternal, "failed to list S3 objects").
			WithContext("bucket", bucketName).
			WithContext("key_prefix", keyPrefix)
	}

	// Prepare to delete the listed objects
	var objectsToDelete []s3types.ObjectIdentifier

	for _, object := range listObjectsResp.Contents {
		objectsToDelete = append(objectsToDelete, s3types.ObjectIdentifier{
			Key: object.Key,
		})
	}

	// If no objects are found, return
	if len(objectsToDelete) == 0 {
		return nil
	}

	// Delete the object
	_, err = a.S3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(bucketName),
		Delete: &s3types.Delete{
			Objects: objectsToDelete,
			Quiet:   aws.Bool(true),
		},
	})

	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeExternal, "failed to delete S3 objects").
			WithContext("bucket", bucketName).
			WithContext("key_prefix", keyPrefix).
			WithContext("object_count", len(objectsToDelete))
	}

	return nil
}

// parseS3Location parses a string in the following format:
// aws:/bucket-name/path-to-file
func (a *AWSClient) parseS3Location(location string) (string, string) {
	pieces := strings.Split(strings.TrimPrefix(location, "aws:"), "/")
	return pieces[0], filepath.Join(pieces[1:]...)
}
