package user_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app/apikey"
	"github.com/stormkit-io/stormkit-io/src/ce/api/user"
	"github.com/stormkit-io/stormkit-io/src/lib/database/databasetest"
	"github.com/stormkit-io/stormkit-io/src/lib/factory"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stretchr/testify/suite"
)

type UserSHTTPSuite struct {
	suite.Suite
	*factory.Factory
	conn databasetest.TestDB
}

func (s *UserSHTTPSuite) BeforeTest(suiteName, _ string) {
	s.conn = databasetest.InitTx(suiteName)
	s.Factory = factory.New(s.conn)
}

func (s *UserSHTTPSuite) AfterTest(_, _ string) {
	s.conn.CloseTx()
}

func (s *UserSHTTPSuite) Test_WithAPIKey_Success() {
	usr := s.MockUser()
	tkn := &apikey.Token{
		Scope:  apikey.SCOPE_USER,
		UserID: usr.ID,
		Value:  apikey.GenerateTokenValue(),
	}

	apikey.NewStore().AddAPIKey(context.Background(), tkn)

	headers := make(http.Header)
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", tkn.Value))

	res := user.WithAPIKey(func(rc *user.RequestContext) *shttp.Response {
		return shttp.OK()
	})(&shttp.RequestContext{
		Request: &http.Request{
			Header: headers,
		},
	})

	s.Equal(res.Status, http.StatusOK)
}

func (s *UserSHTTPSuite) Test_WithAPIKey_Forbidden() {
	headers := make(http.Header)

	res := user.WithAPIKey(func(rc *user.RequestContext) *shttp.Response {
		return shttp.OK()
	})(&shttp.RequestContext{
		Request: &http.Request{
			Header: headers,
		},
	})

	s.Equal(res.Status, http.StatusForbidden)
}

func TestUserSHTTPSuite(t *testing.T) {
	suite.Run(t, &UserModelSuite{})
}
