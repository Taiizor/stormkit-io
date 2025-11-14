package deployhandlers

import (
	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/deploy"
	"github.com/stormkit-io/stormkit-io/src/lib/errors"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
)

// handlerDeployGet requests the current status for the deploy logs.
func handlerDeployGet(req *app.RequestContext) *shttp.Response {
	id := utils.StringToID(req.Vars()["deploymentId"])
	depl, err := deploy.NewStore().DeploymentByIDWithLogs(req.Context(), id, req.App.ID)

	if err != nil {
		wrappedErr := errors.Wrapf(
			err,
			errors.ErrorTypeDatabase,
			"failed to fetch deployment with id=%d for app_id=%d",
			id,
			req.App.ID,
		)
		return shttp.UnexpectedError(wrappedErr)
	}

	if depl == nil {
		return shttp.NotFound()
	}

	depl.DisplayName = req.App.DisplayName

	return &shttp.Response{
		Data: map[string]any{
			"deploy": depl,
		},
	}
}
