package authwallhandlers_test

import (
	"net/http"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app/authwall/authwallhandlers"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stretchr/testify/suite"
)

type ServicesSuite struct {
	suite.Suite
}

func (s *ServicesSuite) Test_Services() {
	services := shttp.NewRouter().RegisterService(authwallhandlers.Services)

	handlers := []string{
		"DELETE:/auth-wall",
		"GET:/auth-wall",
		"GET:/auth-wall/config",
		"POST:/auth-wall",
		"POST:/auth-wall/config",
		"POST:/auth-wall/login",
	}

	s.Equal(handlers, services.HandlerKeys())
}

func (s *ServicesSuite) Test_EE() {
	services := shttp.NewRouter().RegisterService(authwallhandlers.Services)
	s.NotNil(services)

	// All handlers are EE only
	for k, fn := range services.HandlerFuncs() {
		s.Equal(
			http.StatusPaymentRequired,
			fn(&shttp.RequestContext{}).Status,
			"handler %s should return 402", k,
		)
	}
}

func TestServices(t *testing.T) {
	suite.Run(t, &ServicesSuite{})
}
