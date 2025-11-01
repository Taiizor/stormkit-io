package integrations_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/lib/integrations"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/utils/file"
	"github.com/stretchr/testify/suite"
)

type FilesysSuite struct {
	suite.Suite

	tmpdir string
}

func (s *FilesysSuite) BeforeTest(suiteName, _ string) {
	tmpDir, err := os.MkdirTemp("", "deployment-")

	s.NoError(err)

	s.tmpdir = tmpDir
	clientDir := path.Join(tmpDir, "client")

	s.NoError(os.MkdirAll(clientDir, 0774))
	s.NoError(os.WriteFile(path.Join(clientDir, "index.html"), []byte("Hello world"), 0664))
	s.NoError(file.ZipV2(file.ZipArgs{Source: []string{clientDir}, ZipName: path.Join(tmpDir, "sk-client.zip")}))
	s.NoError(file.ZipV2(file.ZipArgs{Source: []string{clientDir}, ZipName: path.Join(tmpDir, "sk-server.zip")}))
	s.NoError(file.ZipV2(file.ZipArgs{Source: []string{clientDir}, ZipName: path.Join(tmpDir, "sk-api.zip")}))
}

func (s *FilesysSuite) AfterTest(_, _ string) {
	if strings.Contains(s.tmpdir, os.TempDir()) {
		os.RemoveAll(s.tmpdir)
	}
}

func (s *FilesysSuite) Test_Upload() {
	client := integrations.Filesys()
	dist := path.Join(s.tmpdir, "dist")

	s.NoError(os.Mkdir(dist, 0774))

	result, err := client.Upload(integrations.UploadArgs{
		DistDir:       dist,
		AppID:         232,
		EnvID:         591,
		DeploymentID:  50919,
		ClientZip:     path.Join(s.tmpdir, "sk-client.zip"),
		ServerZip:     path.Join(s.tmpdir, "sk-server.zip"),
		ServerHandler: "stormkit-server.js:handler",
		APIZip:        path.Join(s.tmpdir, "sk-api.zip"),
		APIHandler:    "stormkit-api.mjs:handler",
	})

	s.NoError(err)
	s.Greater(result.API.BytesUploaded, int64(100))
	s.Greater(result.Server.BytesUploaded, int64(100))
	s.Greater(result.Client.BytesUploaded, int64(0))
	s.Equal(int64(1), result.Client.FilesUploaded)
	s.Equal(fmt.Sprintf("local:%s/deployment-50919/client", dist), result.Client.Location)
	s.Equal(fmt.Sprintf("local:%s/deployment-50919/server/stormkit-server.js:handler", dist), result.Server.Location)
	s.Equal(fmt.Sprintf("local:%s/deployment-50919/api/stormkit-api.mjs:handler", dist), result.API.Location)
}

func (s *FilesysSuite) Test_DeleteArtifacts() {
	client := integrations.Filesys()

	s.DirExists(s.tmpdir)
	s.NoError(client.DeleteArtifacts(context.Background(), integrations.DeleteArtifactsArgs{APILocation: s.tmpdir}))
	s.NoDirExists(s.tmpdir)
}

func (s *FilesysSuite) Test_GetFile() {
	client := integrations.Filesys()
	filePath := path.Join(s.tmpdir, "client", "index.html")

	file, err := client.GetFile(integrations.GetFileArgs{
		Location: fmt.Sprintf("local:%s", filePath),
	})

	s.NoError(err)

	stat, err := os.Stat(filePath)

	s.NoError(err)
	s.Equal(file.Size, stat.Size())
	s.Equal("Hello world", string(file.Content))
	s.Equal("text/html; charset=utf-8", file.ContentType)
}

func (s *FilesysSuite) Test_Invoke() {
	s.NoError(os.WriteFile(path.Join(s.tmpdir, "index.js"), []byte("module.exports = { my_handler: (req, _, cb) => { return cb(null, { body: 'Method is: ' + req.method }) } }"), 0664))

	client := integrations.Filesys()
	reqURL := &url.URL{}

	result, err := client.Invoke(integrations.InvokeArgs{
		URL:         reqURL,
		ARN:         fmt.Sprintf("local:%s:my_handler", path.Join(s.tmpdir, "index.js")),
		Method:      shttp.MethodPut,
		CaptureLogs: true,
	})

	s.NoError(err)
	s.NotEmpty(result)
	s.Equal("Method is: PUT", string(result.Body))
}

func (s *FilesysSuite) Test_Invoke_WithServerCmd() {
	s.NoError(os.WriteFile(path.Join(s.tmpdir, "index.js"), []byte(`
		const http = require('http');

		// Define the hostname and port
		const hostname = '127.0.0.1';
		const port = process.env.PORT;

		// Create the HTTP server
		const server = http.createServer((req, res) => {
			// Set the response HTTP header with HTTP status and Content type
			res.statusCode = 200;
			res.setHeader('Content-Type', 'text/plain');
			// Send the response body "Hello, World!"
			res.end('Hello, World!\n');
		});

		// Make the server listen on the specified port and hostname
		server.listen(port, hostname);
	`), 0664))

	client := integrations.Filesys()
	defer client.ProcessManager().KillAll()

	reqURL := &url.URL{}
	fileName := path.Join(s.tmpdir, "index.js")

	result, err := client.Invoke(integrations.InvokeArgs{
		URL:          reqURL,
		ARN:          fmt.Sprintf("local:%s:my_handler", fileName),
		Method:       shttp.MethodGet,
		Command:      fmt.Sprintf("node %s", fileName),
		DeploymentID: 1,
		CaptureLogs:  true,
	})

	s.NoError(err)
	s.NotEmpty(result)
	s.Equal("Hello, World!\n", string(result.Body))
}

func TestFilesys(t *testing.T) {
	suite.Run(t, &FilesysSuite{})
}

func BenchmarkInvokeWithServerCommand(b *testing.B) {
	s := new(FilesysSuite)
	s.SetT(&testing.T{})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.BeforeTest("Test_Invoke_WithServerCmd", "")
		b.StartTimer()
		s.Test_Invoke_WithServerCmd()
		b.StopTimer()
		s.AfterTest("Test_Invoke_WithServerCmd", "")
	}
}
