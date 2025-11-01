package apikeyhandlers_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app/apikey"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/apikey/apikeyhandlers"
	"github.com/stormkit-io/stormkit-io/src/ce/api/user/usertest"
	"github.com/stormkit-io/stormkit-io/src/lib/database/databasetest"
	"github.com/stormkit-io/stormkit-io/src/lib/factory"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp/shttptest"
)

type HandlerAPIKeySetSuite struct {
	suite.Suite
	*factory.Factory

	conn databasetest.TestDB
}

func (s *HandlerAPIKeySetSuite) BeforeTest(suiteName, _ string) {
	s.conn = databasetest.InitTx(suiteName)
	s.Factory = factory.New(s.conn)
}

func (s *HandlerAPIKeySetSuite) AfterTest(_, _ string) {
	s.conn.CloseTx()
}

func (s *HandlerAPIKeySetSuite) Test_Success() {
	usr := s.MockUser()
	appl := s.MockApp(usr)
	env := s.MockEnv(appl)

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(apikeyhandlers.Services).Router().Handler(),
		shttp.MethodPost,
		"/api-keys",
		map[string]string{
			"appId": appl.ID.String(),
			"envId": env.ID.String(),
			"name":  "Default",
		},
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	data := apikey.Token{}

	s.NoError(json.Unmarshal([]byte(response.String()), &data))
	s.Equal(http.StatusCreated, response.Code)

	s.Equal(appl.ID, data.AppID)
	s.Equal(env.ID, data.EnvID)
	s.Equal(apikey.SCOPE_ENV, data.Scope)
	s.Equal("Default", data.Name)
	s.True(strings.HasPrefix(data.Value, "SK_"))
}

func (s *HandlerAPIKeySetSuite) Test_InvalidScope() {
	usr := s.MockUser()
	appl := s.MockApp(usr)
	env := s.MockEnv(appl)

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(apikeyhandlers.Services).Router().Handler(),
		shttp.MethodPost,
		"/api-keys",
		map[string]string{
			"appId": appl.ID.String(),
			"envId": env.ID.String(),
			"scope": "invalid_scope",
		},
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	data := map[string]string{}

	s.NoError(json.Unmarshal([]byte(response.String()), &data))
	s.Equal(http.StatusBadRequest, response.Code)
	s.Equal(data["error"], "Invalid scope. Allowes scopes are: team, env, user, app, admin")
}

func (s *HandlerAPIKeySetSuite) Test_InvalidName() {
	usr := s.MockUser()
	appl := s.MockApp(usr)
	env := s.MockEnv(appl)

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(apikeyhandlers.Services).Router().Handler(),
		shttp.MethodPost,
		"/api-keys",
		map[string]string{
			"appId": appl.ID.String(),
			"envId": env.ID.String(),
			"scope": "env",
		},
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	data := map[string]string{}

	s.NoError(json.Unmarshal([]byte(response.String()), &data))
	s.Equal(http.StatusBadRequest, response.Code)
	s.Equal(data["error"], "Key name is a required field.")
}

func TestHandlerAPIKeySet(t *testing.T) {
	suite.Run(t, &HandlerAPIKeySetSuite{})
}
