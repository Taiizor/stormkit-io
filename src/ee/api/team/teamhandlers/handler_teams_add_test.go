package teamhandlers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ce/api/admin"
	"github.com/stormkit-io/stormkit-io/src/ce/api/user/usertest"
	"github.com/stormkit-io/stormkit-io/src/ee/api/team"
	"github.com/stormkit-io/stormkit-io/src/ee/api/team/teamhandlers"
	"github.com/stormkit-io/stormkit-io/src/lib/database/databasetest"
	"github.com/stormkit-io/stormkit-io/src/lib/factory"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp/shttptest"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
	"github.com/stretchr/testify/suite"
)

type HandlerTeamsAddSuite struct {
	suite.Suite
	*factory.Factory
	conn databasetest.TestDB
}

func (s *HandlerTeamsAddSuite) BeforeTest(suiteName, _ string) {
	s.conn = databasetest.InitTx(suiteName)
	s.Factory = factory.New(s.conn)
	admin.SetMockLicense()
}

func (s *HandlerTeamsAddSuite) AfterTest(suiteName, _ string) {
	s.conn.CloseTx()
	admin.ResetMockLicense()
}

func (s *HandlerTeamsAddSuite) TestAddTeam() {
	user := s.MockUser()

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(teamhandlers.Services).Router().Handler(),
		shttp.MethodPost,
		"/team",
		map[string]string{
			"name": "My Awesome Team",
		},
		map[string]string{
			"Authorization": usertest.Authorization(user.ID),
		},
	)

	createdTeam := team.Team{}

	s.Equal(http.StatusCreated, response.Code)
	s.NoError(json.Unmarshal([]byte(response.String()), &createdTeam))
	s.Greater(createdTeam.ID, types.ID(1)) // Greater than 1 because there is already the default team
	s.Equal("My Awesome Team", createdTeam.Name)
	s.Equal("my-awesome-team", createdTeam.Slug)
}

func TestHandlerTeamsAdd(t *testing.T) {
	suite.Run(t, &HandlerTeamsAddSuite{})
}
