package mailerhandlers

import (
	"net/http"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/mailer"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
)

func HandlerMailerConfigGet(req *app.RequestContext) *shttp.Response {
	config, err := mailer.Store().Config(req.Context(), req.EnvID)

	if err != nil {
		return shttp.Error(err)
	}

	return &shttp.Response{
		Status: http.StatusOK,
		Data: map[string]any{
			"config": config,
		},
	}
}
