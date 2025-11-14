package file

import (
	stderrors "errors"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/stormkit-io/stormkit-io/src/lib/errors"
)

// File represents an uploaded file.
type File struct {
	// The file struct returned by FormFile.
	file multipart.File

	// Header is the multipart file header returned by FormFile function.
	feader *multipart.FileHeader

	// The path in which the file resides in the file disk.
	Path string

	Permissions os.FileMode
}

// FromRequest returns a new file from the given http request.
func FromRequest(formKey string, r *http.Request) (*File, error) {
	file, header, err := r.FormFile(formKey)

	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeValidation, "failed to get file from form field: "+formKey)
	}

	return &File{
		file:        file,
		feader:      header,
		Permissions: 0666,
	}, nil
}

// New returns a new file instance.
func New(path string) *File {
	return &File{
		Path:        path,
		Permissions: 0666,
	}
}

// WriteContent writes the given source to the file.
func (f *File) WriteContent(src io.Reader) error {
	buf, err := os.OpenFile(f.Path, os.O_WRONLY|os.O_CREATE, f.Permissions)

	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to open file for writing: "+f.Path)
	}

	if _, err = io.Copy(buf, src); err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to write content to file: "+f.Path)
	}

	return nil
}

// SaveToDisk saves a file to disk. The uploaded files will be stored
// in a local tmp folder.
func (f *File) SaveToDisk(destination string) error {
	f.Path = destination
	buf, err := os.OpenFile(f.Path, os.O_WRONLY|os.O_CREATE, f.Permissions)

	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to open destination file: "+destination)
	}

	// Close the buffer when done, and remove the file if copy fails for a reason.
	defer buf.Close()

	if _, err = io.Copy(buf, f.file); err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to save file to disk: "+destination)
	}

	return nil
}

// Remove deletes the file from the filesystem.
func (f *File) Remove() {
	if f.Path != "" {
		os.Remove(f.Path)
	}
}

// CreateFiles iterates on the list of files and creates them in the given location.
// The `files` arguments is a map of strings where the key represents the file name
// including the relative path to the `location` field and value is the file content.
func CreateFiles(files map[string]string, location string) ([]*File, error) {
	var ret = []*File{}

	for file, content := range files {
		var err error
		permissions := fs.FileMode(0777)

		if strings.Contains(file, "/") {
			err = os.MkdirAll(filepath.Join(location, filepath.Dir(file)), permissions)

			if err != nil && !stderrors.Is(err, os.ErrExist) {
				return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to create directory for file: "+file)
			}
		}

		filePath := filepath.Join(location, file)
		err = os.WriteFile(filePath, []byte(content), permissions)

		if err != nil {
			return nil, errors.Wrap(err, errors.ErrorTypeInternal, "failed to write file: "+filePath)
		}

		ret = append(ret, &File{
			Path:        filePath,
			Permissions: permissions,
		})
	}

	return ret, nil
}

// Copy copies a file from src to dst.
func Copy(src, dest string, mode fs.FileMode) error {
	data, err := os.ReadFile(src)

	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to read source file: "+src)
	}

	if err := os.WriteFile(dest, data, mode); err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to write destination file: "+dest)
	}

	return nil
}

// Symlink creates a new symlink by making use of `ln` command.
func Symlink(src, dest string, workdir ...string) error {
	cmd := exec.Command("ln", "-s", src, dest)

	if len(workdir) > 0 {
		cmd.Dir = workdir[0]
	}

	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	cmd.Env = envVars()

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to create symlink from "+src+" to "+dest)
	}

	return nil
}

// Exists checks whether the given file exists or not.
func Exists(fileName string) bool {
	_, err := os.Stat(fileName)
	return err == nil
}
