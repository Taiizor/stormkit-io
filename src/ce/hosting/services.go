package hosting

import (
	"net/http"
	"strings"

	"github.com/stormkit-io/stormkit-io/src/ce/api/admin"
	"github.com/stormkit-io/stormkit-io/src/lib/config"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
)

// Services sets the handlers for this service.
func Services(r *shttp.Router) *shttp.Service {
	pieces := strings.Split(admin.MustConfig().DomainConfig.Dev, "//")
	devDomain := pieces[0]

	if len(pieces) > 1 {
		devDomain = pieces[1]
	}

	s := r.NewService()
	s.NewEndpoint("/").CatchAll(WithHost(HandlerForward), devDomain)

	return s
}

func WithTimeout(h http.Handler) http.Handler {
	return http.TimeoutHandler(h, config.Get().DbConfigTimeouts.ConnectTimeout, "timeout")
}
