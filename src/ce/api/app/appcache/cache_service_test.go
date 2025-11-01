package appcache_test

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app/appcache"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/buildconf"
	"github.com/stormkit-io/stormkit-io/src/lib/database/databasetest"
	"github.com/stormkit-io/stormkit-io/src/lib/factory"
	"github.com/stormkit-io/stormkit-io/src/lib/rediscache"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
	"github.com/stretchr/testify/suite"
)

type CacheSuite struct {
	suite.Suite
	*factory.Factory

	domains []*buildconf.DomainModel
	user    *factory.MockUser
	app     *factory.MockApp
	env     *factory.MockEnv
	ctx     context.Context

	conn databasetest.TestDB
}

func (s *CacheSuite) SetupSuite() {
	s.conn = databasetest.InitTx("appcache_suite")
	s.Factory = factory.New(s.conn)

	s.ctx = context.Background()
	s.user = s.MockUser()
	s.app = s.MockApp(s.user)
	s.env = s.MockEnv(s.app)

	s.domains = []*buildconf.DomainModel{
		{
			AppID:      s.app.ID,
			EnvID:      s.env.ID,
			Name:       "example.org",
			Verified:   true,
			VerifiedAt: utils.NewUnix(),
		},
		{
			AppID:      s.app.ID,
			EnvID:      s.env.ID,
			Name:       "www.example.org",
			Verified:   true,
			VerifiedAt: utils.NewUnix(),
		},
	}

	s.NoError(buildconf.DomainStore().Insert(context.Background(), s.domains[0]))
	s.NoError(buildconf.DomainStore().Insert(context.Background(), s.domains[1]))
}

func (s *CacheSuite) Test_ResetCache() {
	service := rediscache.Service()
	msgs := []string{}

	s.NoError(service.SubscribeAsync(rediscache.EventInvalidateHostingCache, func(ctx context.Context, payload ...string) {
		msgs = append(msgs, payload...)
	}))

	s.NoError(appcache.Service().Reset(s.env.ID))

	expected := []string{
		// First three comes from the first call
		"example.org",
		fmt.Sprintf(`^%s(?:--\d+)?`, s.app.DisplayName),
		"www.example.org",

		// Last one comes from the second call
		"www.example.org",
	}

	// With filters
	s.NoError(appcache.Service().Reset(0, "www.example.org"))

	s.Eventually(func() bool {
		return slices.Equal(expected, msgs)
	}, 5*time.Second, 100*time.Millisecond)
}

func TestCacheSuite(t *testing.T) {
	suite.Run(t, &CacheSuite{})
}
