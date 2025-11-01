package deployhandlers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ce/api/user/usertest"
	"github.com/stormkit-io/stormkit-io/src/lib/config"
	"github.com/stretchr/testify/suite"
	"gopkg.in/guregu/null.v3"

	"github.com/stormkit-io/stormkit-io/src/lib/database/databasetest"
	"github.com/stormkit-io/stormkit-io/src/lib/factory"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp/shttptest"
)

type HandlerDeployGetSuite struct {
	suite.Suite
	*factory.Factory

	conn databasetest.TestDB
}

func (s *HandlerDeployGetSuite) BeforeTest(suiteName, _ string) {
	s.conn = databasetest.InitTx(suiteName)
	s.Factory = factory.New(s.conn)
}

func (s *HandlerDeployGetSuite) AfterTest(_, _ string) {
	s.conn.CloseTx()
}

func (s *HandlerDeployGetSuite) Test_Success() {
	usr := s.MockUser()
	appl := s.MockApp(usr)
	env := s.MockEnv(appl)
	depl := s.MockDeployment(env, map[string]any{
		"ExitCode": null.NewInt(100, true),
	})

	config.Get().Deployer.Service = "github"

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(Services).Router().Handler(),
		shttp.MethodGet,
		fmt.Sprintf("/app/%s/deploy/%s", appl.ID.String(), depl.ID.String()),
		nil,
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	s.Equal(http.StatusOK, response.Code)

	res := response.String()

	s.Contains(res, fmt.Sprintf(`"appId":"%s"`, appl.ID.String()))
	s.Contains(res, fmt.Sprintf(`"id":"%s"`, depl.ID.String()))
	s.Contains(res, `"NODE_ENV":"production"`)
	s.Contains(res, `"exit":100`)
	s.Contains(res, `"isRunning":false`)
	s.NotContains(res, `"preview"`) // exit code 100
}

func (s *HandlerDeployGetSuite) Test_Fail404() {
	usr := s.MockUser()
	usr2 := s.MockUser()
	appl := s.MockApp(usr)
	depl := s.MockDeployment(nil)

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(Services).Router().Handler(),
		shttp.MethodGet,
		fmt.Sprintf("/app/%s/deploy/%s", appl.ID.String(), depl.ID.String()),
		nil,
		map[string]string{
			"Authorization": usertest.Authorization(usr2.ID),
		},
	)

	s.Equal(http.StatusNotFound, response.Code)
}

func TestHandlerDeployGet(t *testing.T) {
	suite.Run(t, &HandlerDeployGetSuite{})
}
