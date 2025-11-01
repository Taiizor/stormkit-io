package adminhandlers_test

import (
	"net/http"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ce/api/admin/adminhandlers"
	"github.com/stormkit-io/stormkit-io/src/ce/api/user/usertest"
	"github.com/stormkit-io/stormkit-io/src/lib/config"
	"github.com/stormkit-io/stormkit-io/src/lib/database/databasetest"
	"github.com/stormkit-io/stormkit-io/src/lib/factory"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp/shttptest"
	"github.com/stretchr/testify/suite"
)

type HandlerLicenseGenerateSuite struct {
	suite.Suite
	*factory.Factory

	conn databasetest.TestDB
}

func (s *HandlerLicenseGenerateSuite) BeforeTest(suiteName, _ string) {
	// Enable cloud mode to make the license generation endpoint available
	config.SetIsStormkitCloud(true)
	s.conn = databasetest.InitTx(suiteName)
	s.Factory = factory.New(s.conn)
}

func (s *HandlerLicenseGenerateSuite) AfterTest(suiteName, _ string) {
	s.conn.CloseTx()
	config.SetIsStormkitCloud(false)
}

func (s *HandlerLicenseGenerateSuite) Test_Success() {
	adminUser := s.MockUser(map[string]any{"IsAdmin": true})

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(adminhandlers.Services).Router().Handler(),
		shttp.MethodPost,
		"/admin/cloud/license",
		map[string]any{
			"seats":       10,
			"description": "Test license for development",
		},
		map[string]string{
			"Authorization": usertest.Authorization(adminUser.ID),
		},
	)

	s.Equal(http.StatusOK, response.Code)
	s.Contains(response.String(), `{"key":"`)
}

func (s *HandlerLicenseGenerateSuite) Test_ExceedsMaximumSeats() {
	adminUser := s.MockUser(map[string]any{"IsAdmin": true})

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(adminhandlers.Services).Router().Handler(),
		shttp.MethodPost,
		"/admin/cloud/license",
		map[string]any{
			"seats":       101,
			"description": "Too many seats",
		},
		map[string]string{
			"Authorization": usertest.Authorization(adminUser.ID),
		},
	)

	s.Equal(http.StatusBadRequest, response.Code)
	s.JSONEq(`{"error":"maximum allowed seats is 100"}`, response.String())
}

func (s *HandlerLicenseGenerateSuite) Test_NonAdmin() {
	nonAdminUser := s.MockUser(map[string]any{"IsAdmin": false})

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(adminhandlers.Services).Router().Handler(),
		shttp.MethodPost,
		"/admin/cloud/license",
		map[string]any{
			"seats":       10,
			"description": "Unauthorized attempt",
		},
		map[string]string{
			"Authorization": usertest.Authorization(nonAdminUser.ID),
		},
	)

	s.Equal(http.StatusUnauthorized, response.Code)
}

func TestHandlerLicenseGenerateSuite(t *testing.T) {
	suite.Run(t, &HandlerLicenseGenerateSuite{})
}
