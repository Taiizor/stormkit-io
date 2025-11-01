package domainhandlers

import (
	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ce/api/user"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
)

// Services sets the handlers for this service.
func Services(r *shttp.Router) *shttp.Service {
	s := r.NewService()

	// Endpoints with environment name.
	s.NewEndpoint("/domains").
		Handler(shttp.MethodGet, "", app.WithApp(HandlerDomainsList, &app.Opts{Env: true})).
		Handler(shttp.MethodPost, "", app.WithApp(HandlerDomainAdd, &app.Opts{Env: true})).
		Handler(shttp.MethodDelete, "", app.WithApp(HandlerDomainDelete, &app.Opts{Env: true})).
		Handler(shttp.MethodGet, "/lookup", app.WithApp(handlerDomainLookup, &app.Opts{Env: true}))

	// Enterprise only
	s.NewEndpoint("/domains").
		Middleware(user.WithEE).
		Handler(shttp.MethodPut, "/cert", app.WithApp(HandlerCertPut, &app.Opts{Env: true})).
		Handler(shttp.MethodDelete, "/cert", app.WithApp(HandlerCertDelete, &app.Opts{Env: true}))

	return s
}
