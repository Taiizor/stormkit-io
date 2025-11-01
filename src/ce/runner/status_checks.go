package runner

import (
	"context"

	"github.com/stormkit-io/stormkit-io/src/lib/utils/sys"
)

type StatusChecks struct {
	workDir  string
	envVars  []string
	reporter *ReporterModel
}

func NewStatusChecks(opts RunnerOpts) *StatusChecks {
	return &StatusChecks{
		workDir:  opts.WorkDir,
		envVars:  opts.Build.EnvVarsRaw,
		reporter: opts.Reporter,
	}
}

func (s *StatusChecks) Run(ctx context.Context, command string) error {
	rep := s.reporter
	rep.AddStep(command)

	cmd := sys.Command(ctx, sys.CommandOpts{
		Name:   "sh",
		Args:   []string{"-c", command},
		Env:    s.envVars,
		Dir:    s.workDir,
		Stdout: rep.File(),
		Stderr: rep.File(),
	})

	err := cmd.Run()

	rep.AddStep("[system] status check passed")

	return err
}
