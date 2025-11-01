package runner_test

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ce/runner"
	"github.com/stretchr/testify/suite"
)

type BuildManagerSuite struct {
	suite.Suite
	config runner.RunnerOpts
}

func (s *BuildManagerSuite) BeforeTest(_, _ string) {
	tmpDir, err := os.MkdirTemp("", "tmp-test-runner-")

	s.NoError(err)

	s.config = runner.RunnerOpts{
		RootDir:  tmpDir,
		Reporter: runner.NewReporter("https://example.com"),
		Repo: runner.RepoOpts{
			Dir: path.Join(tmpDir, "repo"),
		},
	}

	s.NoError(s.config.MkdirAll())
}

func (s *BuildManagerSuite) AfterTest(_, _ string) {
	if strings.Contains(s.config.RootDir, os.TempDir()) {
		s.config.RemoveAll()
	}

	s.config.Reporter.Close(nil, nil, nil)
}

func (s *BuildManagerSuite) Test_Build_Chaining() {
	s.config.Build.BuildCmd = "echo 'hello world' && echo 'hi world'"

	bm := runner.NewBuilder(s.config)
	s.NoError(bm.ExecCommands(context.Background()))

	lines := []string{
		fmt.Sprintf("[sk-step] %s", s.config.Build.BuildCmd),
		"hello world",
		"hi world",
	}

	logs := s.config.Reporter.Logs()

	for _, line := range lines {
		s.Contains(logs, line)
	}
}

func (s *BuildManagerSuite) Test_Build_StaticRepo() {
	bm := runner.NewBuilder(s.config)
	s.NoError(bm.ExecCommands(context.Background()))

	logs := s.config.Reporter.Logs()
	s.Equal("", logs)
}

func TestBuildManagerSuite(t *testing.T) {
	suite.Run(t, &BuildManagerSuite{})
}
