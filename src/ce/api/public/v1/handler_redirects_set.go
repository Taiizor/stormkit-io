package publicapiv1

import (
	"net/http"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/appcache"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/buildconf"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/redirects"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
)

type RedirectsSetRequest struct {
	Redirects []redirects.Redirect `json:"redirects"`
}

func handlerRedirectsSet(req *app.RequestContext) *shttp.Response {
	data := RedirectsSetRequest{}

	if err := req.Post(&data); err != nil {
		return shttp.Error(err)
	}

	store := buildconf.NewStore()
	env, err := store.EnvironmentByID(req.Context(), req.EnvID)

	if err != nil {
		return shttp.Error(err)
	}

	if env == nil {
		return shttp.NotFound()
	}

	env.Data.Redirects = data.Redirects

	if err := store.Update(req.Context(), env); err != nil {
		return shttp.Error(err)
	}

	if err := appcache.Service().Reset(req.EnvID); err != nil {
		return shttp.Error(err)
	}

	return &shttp.Response{
		Status: http.StatusOK,
		Data: map[string]any{
			"redirects": env.Data.Redirects,
		},
	}
}
