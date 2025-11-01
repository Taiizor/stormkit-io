package publicapiv1_test

import (
	"net/http"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app/appcache"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/buildconf"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/redirects"
	publicapiv1 "github.com/stormkit-io/stormkit-io/src/ce/api/public/v1"
	"github.com/stormkit-io/stormkit-io/src/mocks"
	"github.com/stretchr/testify/suite"

	"github.com/stormkit-io/stormkit-io/src/lib/database/databasetest"
	"github.com/stormkit-io/stormkit-io/src/lib/factory"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp/shttptest"
)

type HandlerRedirectsSetSuite struct {
	suite.Suite
	*factory.Factory

	conn             databasetest.TestDB
	mockCacheService *mocks.CacheInterface
}

func (s *HandlerRedirectsSetSuite) SetupSuite() {
	s.mockCacheService = &mocks.CacheInterface{}
}

func (s *HandlerRedirectsSetSuite) BeforeTest(suiteName, _ string) {
	s.conn = databasetest.InitTx(suiteName)
	s.Factory = factory.New(s.conn)
	appcache.DefaultCacheService = s.mockCacheService
}

func (s *HandlerRedirectsSetSuite) AfterTest(_, _ string) {
	s.conn.CloseTx()
	appcache.DefaultCacheService = nil
}

func (s *HandlerRedirectsSetSuite) TestSuccess() {
	reds := []redirects.Redirect{
		{From: "/path", To: "/new-path", Status: http.StatusFound},
		{From: "*", To: "/index.html"},
	}

	usr := s.MockUser()
	app := s.MockApp(usr)
	env := s.MockEnv(app, map[string]any{
		"Data": &buildconf.BuildConf{
			Redirects:     reds,
			RedirectsFile: "/my_file",
		},
	})

	key := s.MockAPIKey(nil, env)

	s.Equal(reds, env.Data.Redirects)
	s.Equal("main", env.Branch)
	s.Equal("/my_file", env.Data.RedirectsFile)

	s.mockCacheService.On("Reset", env.ID).Return(nil)

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(publicapiv1.Services).Router().Handler(),
		shttp.MethodPost,
		"/v1/redirects",
		map[string]any{
			"redirects": []map[string]any{
				{"from": "/path", "to": "/new-path"},
			},
		},
		map[string]string{
			"Authorization": key.Value,
		},
	)

	jsonVal := `{
		"redirects": [
			{ "from":"/path", "to":"/new-path" }
		]
	}`

	s.Equal(http.StatusOK, response.Code)
	s.JSONEq(string(jsonVal), response.String())
}

func TestHandlerRedirectsSet(t *testing.T) {
	suite.Run(t, &HandlerRedirectsSetSuite{})
}
