package redirects_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app/redirects"
	"github.com/stretchr/testify/suite"
)

type RedirectsSuite struct {
	suite.Suite
}

func (s *RedirectsSuite) Test_Redirect_Rewrite() {
	u := &url.URL{
		Scheme: "https",
		Path:   "/my-path",
		Host:   "stormkit.io",
	}

	match := redirects.Match(redirects.MatchArgs{
		URL:      u,
		HostName: "stormkit.io",
		Redirects: []redirects.Redirect{
			{From: "/my-path", To: "/my-new-path"},
		},
	})

	s.NotNil(match)
	s.Equal("/my-new-path", match.Rewrite)
}

func (s *RedirectsSuite) Test_Redirect_Extensions() {
	u := &url.URL{
		Scheme: "https",
		Path:   "/MyAwesomeFile.xsd",
		Host:   "stormkit.io",
	}

	match := redirects.Match(redirects.MatchArgs{
		URL: u,
		Redirects: []redirects.Redirect{
			{From: "/*.xsd", To: "/$1.xsd", Assets: true},
		},
	})

	s.NotNil(match)
	s.Equal("/MyAwesomeFile.xsd", match.Rewrite)
}

func (s *RedirectsSuite) Test_Redirect_TrailingSlash() {
	u := &url.URL{
		Scheme: "https",
		Path:   "/my-path",
		Host:   "stormkit.io",
	}

	match := redirects.Match(redirects.MatchArgs{
		URL:      u,
		HostName: "stormkit.io",
		Redirects: []redirects.Redirect{
			{From: "/my-path", To: "/my-path/", Status: http.StatusFound},
		},
	})

	s.NotNil(match)
	s.Equal("https://stormkit.io/my-path/", match.Redirect)
	s.Equal(http.StatusFound, match.Status)
}

func (s *RedirectsSuite) Test_Redirect_MatchHost() {
	u := &url.URL{
		Scheme: "https",
		Path:   "/my-path",
		Host:   "stormkit.io",
	}

	match := redirects.Match(redirects.MatchArgs{
		URL:      u,
		HostName: "stormkit.io",
		Redirects: []redirects.Redirect{
			{From: "/my-path", To: "/my-path/", Status: http.StatusFound, Hosts: []string{"example.org"}},
		},
	})

	s.Nil(match)

	match = redirects.Match(redirects.MatchArgs{
		URL:      u,
		HostName: "stormkit.io",
		Redirects: []redirects.Redirect{
			{From: "/my-path", To: "/my-path/", Status: http.StatusFound, Hosts: []string{"stormkit.io"}},
		},
	})

	s.NotNil(match)
	s.Equal("https://stormkit.io/my-path/", match.Redirect)
	s.Equal(http.StatusFound, match.Status)
}

func TestRedirects(t *testing.T) {
	suite.Run(t, &RedirectsSuite{})
}
