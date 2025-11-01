package apphandlers_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/apphandlers"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/deploy"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/deployservice"
	"github.com/stormkit-io/stormkit-io/src/lib/database/databasetest"
	"github.com/stormkit-io/stormkit-io/src/lib/factory"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp/shttptest"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
	"github.com/stormkit-io/stormkit-io/src/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AppHooksSuite struct {
	suite.Suite
	*factory.Factory

	conn         databasetest.TestDB
	mockDeployer *mocks.Deployer
}

func (s *AppHooksSuite) BeforeTest(suiteName, _ string) {
	s.conn = databasetest.InitTx(suiteName)
	s.Factory = factory.New(s.conn)
	s.mockDeployer = &mocks.Deployer{}
	s.mockDeployer.On("Deploy", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	deployservice.MockDeployer = s.mockDeployer
}

func (s *AppHooksSuite) AfterTest(_, _ string) {
	s.conn.CloseTx()
	deployservice.MockDeployer = nil
}

func (s *AppHooksSuite) TestSucces() {
	tken := utils.RandomToken(48)
	appl := s.MockApp(nil, map[string]interface{}{
		"DeployTrigger": tken,
	})

	s.MockEnv(appl)

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(apphandlers.Services).Router().Handler(),
		shttp.MethodPost,
		fmt.Sprintf("/hooks/app/%d/deploy/%s/production", appl.ID, tken),
		nil,
		nil,
	)

	a := assert.New(s.T())

	a.Equal(http.StatusOK, response.Code)
	s.mockDeployer.AssertCalled(s.T(), "Deploy",
		mock.Anything, mock.MatchedBy(func(_appl *app.App) bool {
			return a.Equal(appl.ID, _appl.ID)
		}),
		mock.MatchedBy(func(_depl *deploy.Deployment) bool {
			return a.Equal(_depl.CheckoutRepo, "github/svedova/react-minimal") &&
				a.Equal(true, _depl.IsAutoDeploy) &&
				a.Equal(int64(0), _depl.PullRequestNumber.ValueOrZero())
		}))
}

func (s *AppHooksSuite) TestInvalidToken() {
	tkn := utils.RandomToken(48)
	app := s.MockApp(nil)

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(apphandlers.Services).Router().Handler(),
		shttp.MethodPost,
		fmt.Sprintf("/hooks/app/%d/deploy/%s/production", app.ID, tkn),
		nil,
		nil,
	)

	assert.Equal(s.T(), http.StatusUnauthorized, response.Code)
}

func (s *AppHooksSuite) TestOverwritingParams_MethodPOST() {
	tken := utils.RandomToken(48)
	appl := s.MockApp(nil, map[string]interface{}{
		"DeployTrigger": tken,
	})

	s.MockEnv(nil, map[string]interface{}{
		"Branch":      "main",
		"AutoPublish": true,
	})

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(apphandlers.Services).Router().Handler(),
		shttp.MethodPost,
		fmt.Sprintf("/hooks/app/%d/deploy/%s/production", appl.ID, tken),
		map[string]interface{}{
			"branch":  "my-test-branch",
			"publish": false,
		},
		nil,
	)

	a := assert.New(s.T())

	a.Equal(http.StatusOK, response.Code)
	s.mockDeployer.AssertCalled(s.T(), "Deploy",
		mock.Anything, mock.MatchedBy(func(_appl *app.App) bool {
			return a.Equal(appl.ID, _appl.ID)
		}),
		mock.MatchedBy(func(_depl *deploy.Deployment) bool {
			return a.Equal(_depl.Branch, "my-test-branch") &&
				a.Equal(false, _depl.ShouldPublish)
		}))
}

func (s *AppHooksSuite) TestOverwritingParams_MethodGET() {
	tken := utils.RandomToken(48)
	appl := s.MockApp(nil, map[string]interface{}{
		"DeployTrigger": tken,
	})

	s.MockEnv(nil, map[string]interface{}{
		"Branch":      "main",
		"AutoPublish": true,
	})

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(apphandlers.Services).Router().Handler(),
		shttp.MethodGet,
		fmt.Sprintf("/hooks/app/%d/deploy/%s/production?publish=false&branch=my-test-branch", appl.ID, tken),
		nil,
		nil,
	)

	a := assert.New(s.T())

	a.Equal(http.StatusOK, response.Code)
	s.mockDeployer.AssertCalled(s.T(), "Deploy",
		mock.Anything, mock.MatchedBy(func(_appl *app.App) bool {
			return a.Equal(appl.ID, _appl.ID)
		}),
		mock.MatchedBy(func(_depl *deploy.Deployment) bool {
			return a.Equal(_depl.Branch, "my-test-branch") &&
				a.Equal(false, _depl.ShouldPublish)
		}))
}

func (s *AppHooksSuite) TestOverwritingParams_MethodGET_V2() {
	tken := utils.RandomToken(48)
	appl := s.MockApp(nil, map[string]interface{}{
		"DeployTrigger": tken,
	})

	s.MockEnv(nil, map[string]interface{}{
		"Branch":      "main",
		"AutoPublish": false,
	})

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(apphandlers.Services).Router().Handler(),
		shttp.MethodGet,
		fmt.Sprintf("/hooks/app/%d/deploy/%s/production?publish=true&branch=my-test-branch", appl.ID, tken),
		nil,
		nil,
	)

	a := assert.New(s.T())

	a.Equal(http.StatusOK, response.Code)
	s.mockDeployer.AssertCalled(s.T(), "Deploy",
		mock.Anything, mock.MatchedBy(func(_appl *app.App) bool {
			return a.Equal(appl.ID, _appl.ID)
		}),
		mock.MatchedBy(func(_depl *deploy.Deployment) bool {
			return a.Equal(_depl.Branch, "my-test-branch") &&
				a.Equal(true, _depl.ShouldPublish)
		}))
}

func TestHandlerAppHooks(t *testing.T) {
	suite.Run(t, &AppHooksSuite{})
}
