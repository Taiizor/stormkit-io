package userhandlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ce/api/admin"
	"github.com/stormkit-io/stormkit-io/src/ce/api/user"
	"github.com/stormkit-io/stormkit-io/src/ce/api/user/userhandlers"
	"github.com/stormkit-io/stormkit-io/src/ce/api/user/usertest"
	"github.com/stormkit-io/stormkit-io/src/lib/config"
	"github.com/stormkit-io/stormkit-io/src/lib/database/databasetest"
	"github.com/stormkit-io/stormkit-io/src/lib/factory"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp/shttptest"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
	"github.com/stretchr/testify/suite"
)

type HandlerLicenseGetSuite struct {
	suite.Suite
	*factory.Factory
	conn databasetest.TestDB
}

func (s *HandlerLicenseGetSuite) BeforeTest(suiteName, _ string) {
	s.conn = databasetest.InitTx(suiteName)
	s.Factory = factory.New(s.conn)
	config.SetIsStormkitCloud(true)
}

func (s *HandlerLicenseGetSuite) AfterTest(suiteName, _ string) {
	s.conn.CloseTx()
	config.SetIsStormkitCloud(false)
}

func (s *HandlerLicenseGetSuite) Test_Success_NoLicense() {
	usr := s.MockUser(map[string]any{
		"Metadata": user.UserMeta{
			PackageName: config.PackagePremium,
		},
	})

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(userhandlers.Services).Router().Handler(),
		shttp.MethodGet,
		"/user/license",
		nil,
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	res := map[string]any{}

	s.NoError(json.Unmarshal(response.Byte(), &res))
	s.Equal(http.StatusOK, response.Code)
	s.Empty(res["license"])
}

func (s *HandlerLicenseGetSuite) Test_Success_WithLicense() {
	usr := s.MockUser(map[string]any{
		"Metadata": user.UserMeta{
			PackageName: config.PackageFree,
		},
	})

	// Create a license before-hand
	// Generate a new license using the apiKey.value
	license := admin.NewLicense(admin.NewLicenseArgs{
		Seats: 7,
		Key:   "abcd-defg-lfgh-masd-xzva",
	})

	_, err := s.conn.Exec(
		`INSERT INTO licenses (license_key, license_version, user_id, number_of_seats) VALUES ($1, $2, $3, $4)`,
		utils.EncryptToString(license.Key), license.Version, usr.ID, license.Seats,
	)

	s.NoError(err)

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(userhandlers.Services).Router().Handler(),
		shttp.MethodGet,
		"/user/license",
		nil,
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	s.Equal(http.StatusOK, response.Code)

	res := map[string]any{}

	s.NoError(json.Unmarshal(response.Byte(), &res))
	value, ok := res["license"].(map[string]any)
	expected := fmt.Sprintf("%s:%s", usr.ID.String(), "abcd-defg-lfgh-masd-xzva")

	s.True(ok)
	s.NotNil(value)
	s.NotEmpty(value["raw"])
	s.Equal(float64(7), value["seats"])
	s.Equal(expected, value["raw"])
}

func TestHandlerLicenseGetSuite(t *testing.T) {
	suite.Run(t, &HandlerLicenseGetSuite{})
}
