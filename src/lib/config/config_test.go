package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PackageSuite struct {
	suite.Suite

	appSecret string
}

func (s *PackageSuite) BeforeTest(_, _ string) {
	s.appSecret = "gS9u8RZ*3^7^3*jRfDdnTVv9@rrqqr#5"

	os.Setenv("AWS_REGION", "eu-central-1")
	os.Setenv("STORMKIT_APP_SECRET", s.appSecret)
}

func (s *PackageSuite) AfterTest(_, _ string) {
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("STORMKIT_APP_SECRET")
}

func TestPackages(t *testing.T) {
	suite.Run(t, &PackageSuite{})
}
