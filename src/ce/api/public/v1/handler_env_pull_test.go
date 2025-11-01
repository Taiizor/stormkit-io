package publicapiv1_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app/buildconf"
	publicapiv1 "github.com/stormkit-io/stormkit-io/src/ce/api/public/v1"
	"github.com/stretchr/testify/suite"

	"github.com/stormkit-io/stormkit-io/src/lib/database/databasetest"
	"github.com/stormkit-io/stormkit-io/src/lib/factory"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp/shttptest"
)

type HandlerEnvPullSuite struct {
	suite.Suite
	*factory.Factory

	conn databasetest.TestDB
}

func (s *HandlerEnvPullSuite) BeforeTest(suiteName, _ string) {
	s.conn = databasetest.InitTx(suiteName)
	s.Factory = factory.New(s.conn)
}

func (s *HandlerEnvPullSuite) AfterTest(_, _ string) {
	s.conn.CloseTx()
}

func (s *HandlerEnvPullSuite) TestSuccess() {
	vars := map[string]string{
		"NODE_ENV": "production",
		"TEST_1":   "VALUE_1",
		"TEST_2":   "VALUE_1=VALUE_1",
	}

	env := s.MockEnv(nil, map[string]any{
		"Data": &buildconf.BuildConf{
			Vars: vars,
		},
	})

	key := s.MockAPIKey(nil, env)

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(publicapiv1.Services).Router().Handler(),
		shttp.MethodGet,
		"/v1/env/pull",
		nil,
		map[string]string{
			"Authorization": key.Value,
		},
	)

	jsonVal, err := json.Marshal(vars)
	s.NoError(err)

	s.Equal(http.StatusOK, response.Code)
	s.JSONEq(response.String(), string(jsonVal))
}

func TestHandlerEnvPull(t *testing.T) {
	suite.Run(t, &HandlerEnvPullSuite{})
}
