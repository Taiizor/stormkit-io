package appconf_test

import (
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ce/api/admin"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/appconf"
	"github.com/stormkit-io/stormkit-io/src/lib/database/databasetest"
	"github.com/stormkit-io/stormkit-io/src/lib/factory"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
	"github.com/stretchr/testify/suite"
)

type ShttpSuite struct {
	suite.Suite
	*factory.Factory

	conn databasetest.TestDB
}

func (s *ShttpSuite) BeforeTest(suiteName, _ string) {
	s.conn = databasetest.InitTx(suiteName)
	s.Factory = factory.New(s.conn)
}

func (s *ShttpSuite) AfterTest(_, _ string) {
	s.conn.CloseTx()
}

func (s *ShttpSuite) Test_IsStormkitDev() {
	admin.MustConfig().DomainConfig.Dev = "http://stormkit:8888"

	s.True(appconf.IsStormkitDev("my-app--1868181.stormkit:8888"))
	s.True(appconf.IsStormkitDev("my-app.stormkit:8888"))
	s.False(appconf.IsStormkitDev("my-app.stormkit"))
	s.False(appconf.IsStormkitDev("my-app.com"))
	s.False(appconf.IsStormkitDev("stormkit:8888"))

	admin.MustConfig().DomainConfig.Dev = "http://stormkit.dev"

	s.True(appconf.IsStormkitDev("my-app--1868181.stormkit.dev"))
	s.True(appconf.IsStormkitDev("my-app.stormkit.dev"))
	s.False(appconf.IsStormkitDev("my-app.stormkit.dev.io"))
	s.False(appconf.IsStormkitDev("my-app.com"))
	s.False(appconf.IsStormkitDev("stormkit:8888"))
}

func (s *ShttpSuite) Test_ParseHost() {
	admin.MustConfig().DomainConfig.Dev = "http://stormkit.dev"

	customDomains := []string{
		"dev.app",
		"dev.app.stormkit",
		"stormkit.de",
		"my-app--1.my.domain",
	}

	for _, d := range customDomains {
		s.Equal(appconf.RequestContext{
			DomainName:   d,
			DisplayName:  "",
			EnvName:      "",
			DeploymentID: 0,
			App:          nil,
			User:         nil,
		}, *appconf.ParseHost(d))
	}

	devDomains := []appconf.RequestContext{
		{DomainName: "app.stormkit.dev", DisplayName: "app"},
		{DomainName: "dev.app.stormkit.dev", DisplayName: "dev.app"},
		{DomainName: "app--dev.stormkit.dev", DisplayName: "app", EnvName: "dev"},
		{DomainName: "app--1.stormkit.dev", DisplayName: "app", EnvName: "", DeploymentID: types.ID(1)},
	}

	for _, d := range devDomains {
		s.Equal(appconf.RequestContext{
			DomainName:   d.DomainName,
			DisplayName:  d.DisplayName,
			EnvName:      d.EnvName,
			DeploymentID: d.DeploymentID,
			App:          nil,
			User:         nil,
		}, *appconf.ParseHost(d.DomainName))
	}
}

func (s *ShttpSuite) Test_IsStormkitDevStrict() {
	admin.MustConfig().DomainConfig.Dev = "http://stormkit:8888"
	s.False(appconf.IsStormkitDevStrict("my-app--1868181.stormkit:8888"))
	s.True(appconf.IsStormkitDevStrict("stormkit:8888"))
}

func TestShttpMethods(t *testing.T) {
	suite.Run(t, &ShttpSuite{})
}
