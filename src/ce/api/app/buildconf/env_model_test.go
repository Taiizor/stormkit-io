package buildconf_test

import (
	"fmt"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app/buildconf"
	"github.com/stormkit-io/stormkit-io/src/lib/database/databasetest"
	"github.com/stormkit-io/stormkit-io/src/lib/factory"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stretchr/testify/suite"
)

type EnvModelSuite struct {
	suite.Suite
	*factory.Factory

	conn databasetest.TestDB
}

func (s *EnvModelSuite) BeforeTest(suiteName, _ string) {
	s.conn = databasetest.InitTx(suiteName)
	s.Factory = factory.New(s.conn)
}

func (s *EnvModelSuite) AfterTest(_, _ string) {
	s.conn.CloseTx()
}

func (s *EnvModelSuite) TestConfig_Validation() {
	config := &buildconf.Env{}

	res := shttp.Error(config.Validate())
	exp := fmt.Sprintf(`{"errors":{"branch":"%s","env":"Environment is missing"}}`, buildconf.ErrInvalidBranch.Error())
	s.Equal(res.String(), exp)

	config.Env = "Some invalid env"
	config.Branch = "Valid-Env-1015+=/z"
	res = shttp.Error(config.Validate())
	exp = fmt.Sprintf(`{"errors":{"env":"%s"}}`, buildconf.ErrInvalidEnv.Error())
	s.Equal(res.String(), exp)

	config.Env = "Some-Valid-Env-Name"
	config.Branch = "I'm invalid"
	res = shttp.Error(config.Validate())
	exp = fmt.Sprintf(`{"errors":{"branch":"%s"}}`, buildconf.ErrInvalidBranch.Error())
	s.Equal(res.String(), exp)
}

func TestEnvModelSuite(t *testing.T) {
	suite.Run(t, &EnvModelSuite{})
}
