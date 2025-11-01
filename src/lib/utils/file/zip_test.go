package file_test

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/lib/utils/file"
	"github.com/stretchr/testify/suite"
)

type ZipSuite struct {
	suite.Suite
	tmpDir string
}

func (s *ZipSuite) BeforeTest(_, _ string) {
	var err error
	s.tmpDir, err = os.MkdirTemp("", "zip_test")
	s.NoError(err)
}

func (s *ZipSuite) AfterTest(_, _ string) {
	s.NoError(os.RemoveAll(s.tmpDir))
}

func (s *ZipSuite) createZipWithFiles(files map[string][]byte) []byte {
	// Create a buffer to hold the zip content
	var buf bytes.Buffer

	// Create a new zip writer
	zipWriter := zip.NewWriter(&buf)

	for fileName, fileContent := range files {
		zf, err := zipWriter.Create(fileName)
		s.NoError(err)
		_, err = zf.Write(fileContent)
		s.NoError(err)
	}

	// Close the zip writer to finalize the zip content
	s.NoError(zipWriter.Close())

	// Return the zip content as a byte slice
	return buf.Bytes()
}

func (s *ZipSuite) TestUnzip_ValidZipFile() {
	files := map[string][]byte{
		"index.html":         []byte("Hello World"),
		"my/folder/file.txt": []byte("This is a test file."),
	}

	// Create a valid zip file
	zipContent := s.createZipWithFiles(files)
	zipFile := filepath.Join(s.tmpDir, "test.zip")

	s.NoError(os.WriteFile(zipFile, zipContent, 0644))

	// Unzip the file
	s.NoError(os.MkdirAll(filepath.Join(s.tmpDir, "output"), 0755))
	destDir := filepath.Join(s.tmpDir, "output")
	s.NoError(file.Unzip(file.UnzipOpts{zipFile, destDir, false}))

	// Verify the unzipped content
	unzippedFile := filepath.Join(destDir, "index.html")
	content, err := os.ReadFile(unzippedFile)
	s.NoError(err)
	s.Equal([]byte("Hello World"), content)
}

func (s *ZipSuite) TestUnzip_ZipSlipVulnerability() {
	files := map[string][]byte{
		"../index.html": []byte("Hello World"),
	}

	// Create a valid zip file
	zipContent := s.createZipWithFiles(files)
	zipFile := filepath.Join(s.tmpDir, "test-invalid.zip")

	s.NoError(os.WriteFile(zipFile, zipContent, 0644))

	// Unzip the file
	s.NoError(os.MkdirAll(filepath.Join(s.tmpDir, "output"), 0755))
	destDir := filepath.Join(s.tmpDir, "output")
	err := file.Unzip(file.UnzipOpts{zipFile, destDir, false})
	s.Error(err)
}

func TestZipSuite(t *testing.T) {
	suite.Run(t, &ZipSuite{})
}
