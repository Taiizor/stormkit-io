package volumes

import (
	"fmt"
	"io"
	"mime/multipart"
	"path"

	"github.com/stormkit-io/stormkit-io/src/ce/api/admin"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
)

const (
	FileSys    = "filesys"
	AWSS3      = "aws:s3"
	AlibabaOSS = "alibaba:oss"
	HCloudOSS  = "hcloud:oss"
)

type File struct {
	ID        types.ID
	EnvID     types.ID
	Name      string // The file name with the maintained structure (e.g. test-folder/file.txt)
	Path      string // The absolute path to the file including relative path (/shared/volumes/test-folder)
	Size      int64
	IsPublic  bool
	CreatedAt utils.Unix
	UpdatedAt utils.Unix
	Metadata  utils.Map
}

// FullPath returns the absolute path of the file.
func (f *File) FullPath() string {
	return path.Join(f.Path, path.Base(f.Name))
}

// PublicLink returns the link to access the file.
func (f *File) PublicLink() string {
	token := utils.EncryptToString(f.ID.String() + ":" + f.EnvID.String())
	return admin.MustConfig().ApiURL(fmt.Sprintf("/volumes/file/%s", token))
}

type UploadArgs struct {
	AppID              types.ID
	EnvID              types.ID
	FileHeader         FileHeader
	ContentDisposition map[string]string
}

// Upload a file to the destination specified by args.MountType.
func Upload(vc *admin.VolumesConfig, args UploadArgs) (*File, error) {
	file, err := args.FileHeader.Open()

	if err != nil {
		return nil, err
	}

	defer file.Close()

	fileName := utils.GetString(args.ContentDisposition["filename"], args.FileHeader.Name())
	filePath := constructFilePath(args.AppID, args.EnvID)

	opts := uploadArgs{
		file:        file,
		size:        args.FileHeader.Size(),
		envID:       args.EnvID,
		dstFileName: fileName,
		dstFilePath: filePath,
	}

	if vc.MountType == FileSys {
		return clientFilesys(vc).upload(opts)
	} else if vc.MountType == AWSS3 {
		return clientAWS(vc).upload(opts)
	}

	return nil, nil
}

// Download downloads a file from the source.
func Download(vc *admin.VolumesConfig, file *File) (io.ReadSeeker, error) {
	if vc.MountType == FileSys {
		return clientFilesys(vc).download(file)
	} else if vc.MountType == AWSS3 {
		return clientAWS(vc).download(file)
	}

	return nil, nil
}

// Upload a file to the destination specified by args.MountType.
// This function returns a list of files that are successfully removed.
// When encountered an error other than os.IsNotExist, returns immediately.
func RemoveFiles(vc *admin.VolumesConfig, files []*File) ([]*File, error) {
	if vc.MountType == FileSys {
		return clientFilesys(vc).removeFiles(files)
	} else if vc.MountType == AWSS3 {
		return clientAWS(vc).removeFiles(files)
	}

	return nil, nil
}

func constructFilePath(appID, envID types.ID) string {
	return path.Join(fmt.Sprintf("a%se%s", appID, envID))
}

type FileHeader interface {
	Open() (multipart.File, error)
	Name() string
	Size() int64
}

type fileHeader struct {
	original *multipart.FileHeader
}

func (fh *fileHeader) Open() (multipart.File, error) {
	return fh.original.Open()
}

func (fh *fileHeader) Size() int64 {
	return fh.original.Size
}

func (fh *fileHeader) Name() string {
	return fh.original.Filename
}

func FromFileHeader(file *multipart.FileHeader) FileHeader {
	return &fileHeader{
		original: file,
	}
}
