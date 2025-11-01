package deployhandlers_test

import (
	"net/http"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/deploy"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/deployservice"
	"github.com/stormkit-io/stormkit-io/src/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gopkg.in/guregu/null.v3"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app/deploy/deployhandlers"
	"github.com/stormkit-io/stormkit-io/src/ce/api/user/usertest"
	"github.com/stormkit-io/stormkit-io/src/lib/database/databasetest"
	"github.com/stormkit-io/stormkit-io/src/lib/factory"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp/shttptest"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
)

type HandlerDeploymentRestartSuite struct {
	suite.Suite
	*factory.Factory

	conn         databasetest.TestDB
	mockDeployer *mocks.Deployer
}

func (s *HandlerDeploymentRestartSuite) BeforeTest(suiteName, _ string) {
	s.conn = databasetest.InitTx(suiteName)
	s.Factory = factory.New(s.conn)
	s.mockDeployer = &mocks.Deployer{}
	s.mockDeployer.On("Deploy", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	deployservice.MockDeployer = s.mockDeployer
}

func (s *HandlerDeploymentRestartSuite) AfterTest(_, _ string) {
	s.conn.CloseTx()
	deployservice.MockDeployer = nil
}

func (s *HandlerDeploymentRestartSuite) Test_Success() {
	usr := s.MockUser()
	appl := s.MockApp(usr)
	env := s.MockEnv(appl)
	depl := s.MockDeployment(env, map[string]any{
		"StoppedAt": utils.NewUnix(),
		"ExitCode":  null.IntFrom(1),
	})

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(deployhandlers.Services).Router().Handler(),
		shttp.MethodPost,
		"/app/deploy/restart",
		map[string]any{
			"appId":        appl.ID.String(),
			"envId":        env.ID.String(),
			"deploymentId": depl.ID.String(),
		},
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	s.Equal(http.StatusOK, response.Code)

	s.mockDeployer.AssertCalled(s.T(), "Deploy",
		mock.Anything, mock.MatchedBy(func(a *app.App) bool {
			return s.Equal(appl.ID, a.ID)
		}),
		mock.MatchedBy(func(d *deploy.Deployment) bool {
			return s.Equal(d.CheckoutRepo, depl.CheckoutRepo)
		}),
	)
}

func (s *HandlerDeploymentRestartSuite) Test_BadRequest_OnlyFailedDeployments() {
	usr := s.MockUser()
	appl := s.MockApp(usr)
	env := s.MockEnv(appl)
	depl := s.MockDeployment(env)

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(deployhandlers.Services).Router().Handler(),
		shttp.MethodPost,
		"/app/deploy/restart",
		map[string]any{
			"appId":        appl.ID.String(),
			"envId":        env.ID.String(),
			"deploymentId": depl.ID.String(),
		},
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	expected := `{ "error" : "Only failed deployments can be restarted." }`

	s.Equal(http.StatusBadRequest, response.Code)
	s.JSONEq(response.String(), expected)

	s.mockDeployer.AssertNotCalled(s.T(), "Deploy")
}

func (s *HandlerDeploymentRestartSuite) Test_BadRequest() {
	usr := s.MockUser()
	appl := s.MockApp(usr)
	env := s.MockEnv(appl)

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(deployhandlers.Services).Router().Handler(),
		shttp.MethodPost,
		"/app/deploy/restart",
		map[string]any{
			"appId": appl.ID.String(),
			"envId": env.ID.String(),
		},
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	s.Equal(response.Code, http.StatusBadRequest)

	s.mockDeployer.AssertNotCalled(s.T(), "Deploy")
}

func TestHandlerDeploymentRestart(t *testing.T) {
	suite.Run(t, &HandlerDeploymentRestartSuite{})
}
