package apikey_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app/apikey"
	"github.com/stormkit-io/stormkit-io/src/lib/database/databasetest"
	"github.com/stormkit-io/stormkit-io/src/lib/factory"
)

type StoreSuite struct {
	suite.Suite
	*factory.Factory

	conn databasetest.TestDB
	app  *factory.MockApp
	env  *factory.MockEnv
}

func (s *StoreSuite) SetupSuite() {
	s.conn = databasetest.InitTx("apikey_store_suite")
	s.Factory = factory.New(s.conn)

	s.app = s.MockApp(nil)
	s.env = s.MockEnv(s.app)
}

func (s *StoreSuite) TearDownSuite() {
	s.conn.CloseTx()
}

func (s *StoreSuite) TestAddAPIKey_Success() {
	key := &apikey.Token{
		Name:  "Something",
		Scope: apikey.SCOPE_ENV,
		AppID: s.app.ID,
		EnvID: s.env.ID,
	}

	err := apikey.NewStore().AddAPIKey(context.Background(), key)

	s.NoError(err)
	s.True(key.ID > 0)
}

func (s *StoreSuite) TestAddAPIKey_InvalidScope() {
	key := &apikey.Token{
		Name:  "Something",
		Scope: "something-else",
	}

	err := apikey.NewStore().AddAPIKey(context.Background(), key)

	s.Error(err)
	s.Equal("scope is invalid", err.Error())
}

func TestStoreSuite(t *testing.T) {
	suite.Run(t, &StoreSuite{})
}
