package jobs_test

import (
	"context"
	"testing"
	"time"

	jobs "github.com/stormkit-io/stormkit-io/src/ce/workerserver"
	"github.com/stormkit-io/stormkit-io/src/lib/database/databasetest"
	"github.com/stormkit-io/stormkit-io/src/lib/factory"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
	"github.com/stretchr/testify/suite"
)

type JobEnvironmentsSuite struct {
	suite.Suite
	*factory.Factory
	conn databasetest.TestDB
}

func (s *JobEnvironmentsSuite) BeforeTest(suiteName, _ string) {
	s.conn = databasetest.InitTx(suiteName)
	s.Factory = factory.New(s.conn)
}

func (s *JobEnvironmentsSuite) AfterTest(suiteName, _ string) {
	s.conn.CloseTx()
}

func (s *JobEnvironmentsSuite) Test_RemoveStateEnvironments() {
	app := s.MockApp(nil)
	env := s.MockEnv(app, map[string]any{
		"DeletedAt": utils.Unix{
			Time:  time.Now(),
			Valid: true,
		},
	})

	s.MockDeployments(10, env)

	env2 := s.MockEnv(app, map[string]any{
		"DeletedAt": utils.Unix{
			Time:  time.Now(),
			Valid: true,
		},
	})

	s.MockDeployments(10, env2)

	err := jobs.RemoveStaleEnvironments(context.TODO())
	s.NoError(err)

	tmp, _ := s.conn.Prepare("SELECT deleted_at FROM deployments;")
	res, _ := tmp.Query()

	defer res.Close()

	for res.Next() {
		var deletedAt utils.Unix
		err := res.Scan(&deletedAt)
		s.NoError(err)
		s.NotEmpty(deletedAt)
	}
}

func TestJobEnvironmentsSuite(t *testing.T) {
	suite.Run(t, &JobEnvironmentsSuite{})
}
