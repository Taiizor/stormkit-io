package deployhandlers_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app/deploy"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/deploy/deployhandlers"
	"github.com/stormkit-io/stormkit-io/src/ce/api/user/usertest"
	"github.com/stormkit-io/stormkit-io/src/lib/database/databasetest"
	"github.com/stormkit-io/stormkit-io/src/lib/factory"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp/shttptest"
	"github.com/stretchr/testify/suite"
	"gopkg.in/guregu/null.v3"
)

type HandlerDeploymentsSuite struct {
	suite.Suite
	*factory.Factory

	conn databasetest.TestDB
}

func (s *HandlerDeploymentsSuite) BeforeTest(suiteName, _ string) {
	s.conn = databasetest.InitTx(suiteName)
	s.Factory = factory.New(s.conn)
}

func (s *HandlerDeploymentsSuite) AfterTest(_, _ string) {
	s.conn.CloseTx()
}

func (s *HandlerDeploymentsSuite) Test_WithoutFilters() {
	usr := s.MockUser()
	appl := s.MockApp(usr)
	env := s.MockEnv(appl)
	_ = s.MockDeployments(3, env, map[string]any{
		"Published": []deploy.PublishedInfo{
			{Percentage: 70, EnvID: env.ID},
			{Percentage: 30, EnvID: env.ID},
		},
	})

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(deployhandlers.Services).Router().Handler(),
		shttp.MethodPost,
		"/app/deployments",
		map[string]any{
			"appId": appl.ID.String(),
		},
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	jsonResponse := map[string]any{}

	s.NoError(json.Unmarshal(response.Body.Bytes(), &jsonResponse))

	deployments := jsonResponse["deploys"].([]any)
	hasNextPage := jsonResponse["hasNextPage"].(bool)

	str := response.String()
	envID := env.ID.String()

	s.Len(deployments, 3)
	s.Equal(false, hasNextPage)
	s.Contains(str, fmt.Sprintf(`"appId":"%s"`, appl.ID.String()))
	s.Contains(str, fmt.Sprintf(`"published":[{"envId":"%s","percentage":70},{"envId":"%s","percentage":30}]`, envID, envID))
}

func (s *HandlerDeploymentsSuite) Test_WithFiltersAndNoResults() {
	usr := s.MockUser()
	appl := s.MockApp(usr)
	env := s.MockEnv(appl)

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(deployhandlers.Services).Router().Handler(),
		shttp.MethodPost,
		"/app/deployments",
		map[string]interface{}{
			"appId":     appl.ID.String(),
			"envId":     env.ID.String(),
			"branch":    "master",
			"status":    true,
			"published": true,
		},
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	jsonResponse := map[string]any{}

	s.NoError(json.Unmarshal(response.Body.Bytes(), &jsonResponse))

	deployments := jsonResponse["deploys"].([]any)
	hasNextPage := jsonResponse["hasNextPage"].(bool)

	s.Empty(deployments)
	s.Equal(hasNextPage, false)
}

func (s *HandlerDeploymentsSuite) Test_WithFiltersAndSomeResults() {
	usr := s.MockUser()
	appl := s.MockApp(usr)
	env := s.MockEnv(appl)
	_ = s.MockDeployments(3, env, map[string]any{
		"ExitCode": null.NewInt(1, true),
	}, map[string]any{
		"ExitCode": null.NewInt(1, true),
	}, map[string]any{
		"ExitCode": null.NewInt(0, true),
	})

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(deployhandlers.Services).Router().Handler(),
		shttp.MethodPost,
		"/app/deployments",
		map[string]any{
			"appId":  appl.ID.String(),
			"status": false,
		},
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	jsonResponse := map[string]any{}

	s.NoError(json.Unmarshal(response.Body.Bytes(), &jsonResponse))

	deployments := jsonResponse["deploys"].([]any)
	hasNextPage := jsonResponse["hasNextPage"].(bool)
	responseStr := response.String()

	s.Len(deployments, 2)
	s.Equal(hasNextPage, false)
	s.Contains(responseStr, fmt.Sprintf(`"appId":"%s"`, appl.ID.String()))
}

func TestHandlerDeployments(t *testing.T) {
	suite.Run(t, &HandlerDeploymentsSuite{})
}
