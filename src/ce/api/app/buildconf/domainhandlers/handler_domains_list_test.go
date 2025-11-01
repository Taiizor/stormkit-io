package domainhandlers_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app/buildconf"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/buildconf/domainhandlers"
	"github.com/stormkit-io/stormkit-io/src/ce/api/user/usertest"
	"github.com/stormkit-io/stormkit-io/src/lib/database/databasetest"
	"github.com/stormkit-io/stormkit-io/src/lib/factory"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp/shttptest"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
	"github.com/stretchr/testify/suite"
	"gopkg.in/guregu/null.v3"
)

type HandlerDomainsList struct {
	suite.Suite
	*factory.Factory
	conn databasetest.TestDB
}

func (s *HandlerDomainsList) BeforeTest(suiteName, _ string) {
	s.conn = databasetest.InitTx(suiteName)
	s.Factory = factory.New(s.conn)
}

func (s *HandlerDomainsList) AfterTest(_, _ string) {
	s.conn.CloseTx()
}

func (s *HandlerDomainsList) Test_Success_WithoutFilters() {
	usr := s.Factory.MockUser()
	app := s.Factory.MockApp(usr, nil)
	env := s.Factory.MockEnv(app)

	d1 := &buildconf.DomainModel{
		AppID:      app.ID,
		EnvID:      env.ID,
		Name:       "example.org",
		Verified:   true,
		VerifiedAt: utils.NewUnix(),
	}

	d2 := &buildconf.DomainModel{
		AppID:    app.ID,
		EnvID:    env.ID,
		Name:     "my.example.org",
		Token:    null.NewString("my-token", true),
		Verified: false,
	}

	s.NoError(buildconf.DomainStore().Insert(context.Background(), d1))
	s.NoError(buildconf.DomainStore().Insert(context.Background(), d2))

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(domainhandlers.Services).Router().Handler(),
		shttp.MethodGet,
		fmt.Sprintf("/domains?appId=%s&envId=%s", app.ID.String(), env.ID.String()),
		nil,
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	expected := fmt.Sprintf(`{
		"domains": [
			{ "id": "%d", "domainName": "example.org", "verified": true, "token": "", "customCert": null, "lastPing": null },
			{ "id": "%d", "domainName": "my.example.org", "verified": false, "token": "my-token", "customCert": null, "lastPing": null }
		],
		"pagination": {
			"hasNextPage": false
		}
	}`, d1.ID, d2.ID)

	s.Equal(http.StatusOK, response.Code)
	s.JSONEq(expected, response.String())
}

func (s *HandlerDomainsList) Test_Success_Pagination() {
	usr := s.Factory.MockUser()
	app := s.Factory.MockApp(usr, nil)
	env := s.Factory.MockEnv(app)

	domainhandlers.DefaultDomainsLimit = 1

	defer func() {
		domainhandlers.DefaultDomainsLimit = 100
	}()

	d1 := &buildconf.DomainModel{
		AppID:      app.ID,
		EnvID:      env.ID,
		Name:       "example.org",
		Verified:   true,
		VerifiedAt: utils.NewUnix(),
	}

	d2 := &buildconf.DomainModel{
		AppID:    app.ID,
		EnvID:    env.ID,
		Name:     "my.example.org",
		Token:    null.NewString("my-token", true),
		Verified: false,
	}

	s.NoError(buildconf.DomainStore().Insert(context.Background(), d1))
	s.NoError(buildconf.DomainStore().Insert(context.Background(), d2))

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(domainhandlers.Services).Router().Handler(),
		shttp.MethodGet,
		fmt.Sprintf("/domains?appId=%s&envId=%s", app.ID.String(), env.ID.String()),
		nil,
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	expected := fmt.Sprintf(`{
		"domains": [
			{ "id": "%d", "domainName": "example.org", "verified": true, "token": "", "customCert": null, "lastPing": null }
		],
		"pagination": {
			"hasNextPage": true,
			"afterId": "%d"
		}
	}`, d1.ID, d1.ID)

	s.Equal(http.StatusOK, response.Code)
	s.JSONEq(expected, response.String())

	response = shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(domainhandlers.Services).Router().Handler(),
		shttp.MethodGet,
		fmt.Sprintf(
			"/domains?appId=%s&envId=%s&afterId=%d",
			app.ID.String(),
			env.ID.String(),
			d1.ID,
		),
		nil,
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	expected = fmt.Sprintf(`{
		"domains": [
			{ "id": "%d", "domainName": "my.example.org", "verified": false, "token": "my-token", "customCert": null, "lastPing": null }
		],
		"pagination": {
			"hasNextPage": false
		}
	}`, d2.ID)

	s.Equal(http.StatusOK, response.Code)
	s.JSONEq(expected, response.String())
}

func (s *HandlerDomainsList) Test_Success_WithFilters() {
	usr := s.Factory.MockUser()
	app := s.Factory.MockApp(usr, nil)
	env := s.Factory.MockEnv(app)

	d1 := &buildconf.DomainModel{
		AppID:      app.ID,
		EnvID:      env.ID,
		Name:       "example.org",
		Verified:   true,
		VerifiedAt: utils.NewUnix(),
	}

	d2 := &buildconf.DomainModel{
		AppID:    app.ID,
		EnvID:    env.ID,
		Name:     "my.example.org",
		Token:    null.NewString("my-token", true),
		Verified: false,
	}

	s.NoError(buildconf.DomainStore().Insert(context.Background(), d1))
	s.NoError(buildconf.DomainStore().Insert(context.Background(), d2))

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(domainhandlers.Services).Router().Handler(),
		shttp.MethodGet,
		fmt.Sprintf("/domains?appId=%s&envId=%s&verified=true", app.ID.String(), env.ID.String()),
		nil,
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	expected := fmt.Sprintf(`{
		"domains": [
			{ "id": "%d", "domainName": "example.org", "verified": true, "token": "", "customCert": null, "lastPing": null }
		],
		"pagination": {
			"hasNextPage": false
		}
	}`, d1.ID)

	s.Equal(http.StatusOK, response.Code)
	s.JSONEq(expected, response.String())
}

func (s *HandlerDomainsList) Test_Success_WithFilters_DomainName() {
	usr := s.Factory.MockUser()
	app := s.Factory.MockApp(usr, nil)
	env := s.Factory.MockEnv(app)

	d1 := &buildconf.DomainModel{
		AppID:      app.ID,
		EnvID:      env.ID,
		Name:       "example.org",
		Verified:   true,
		VerifiedAt: utils.NewUnix(),
	}

	d2 := &buildconf.DomainModel{
		AppID:    app.ID,
		EnvID:    env.ID,
		Name:     "my.other.org",
		Token:    null.NewString("my-token", true),
		Verified: false,
	}

	s.NoError(buildconf.DomainStore().Insert(context.Background(), d1))
	s.NoError(buildconf.DomainStore().Insert(context.Background(), d2))

	response := shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(domainhandlers.Services).Router().Handler(),
		shttp.MethodGet,
		fmt.Sprintf("/domains?appId=%s&envId=%s&domainName=xAMp", app.ID.String(), env.ID.String()),
		nil,
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	expected := fmt.Sprintf(`{
		"domains": [
			{ "id": "%d", "domainName": "example.org", "verified": true, "token": "", "customCert": null, "lastPing": null }
		],
		"pagination": {
			"hasNextPage": false
		}
	}`, d1.ID)

	s.Equal(http.StatusOK, response.Code)
	s.JSONEq(expected, response.String())

	response = shttptest.RequestWithHeaders(
		shttp.NewRouter().RegisterService(domainhandlers.Services).Router().Handler(),
		shttp.MethodGet,
		fmt.Sprintf("/domains?appId=%s&envId=%s&domainName=hello", app.ID.String(), env.ID.String()),
		nil,
		map[string]string{
			"Authorization": usertest.Authorization(usr.ID),
		},
	)

	s.Equal(http.StatusOK, response.Code)
	s.JSONEq(`{ "domains": [], "pagination": { "hasNextPage": false }}`, response.String())
}

func TestHandlerDomainsList(t *testing.T) {
	suite.Run(t, &HandlerDomainsList{})
}
