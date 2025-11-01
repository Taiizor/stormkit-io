package deployhandlers

import (
	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/deploy"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/deployservice"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
)

type deployStopRequest struct {
	DeploymentID types.ID `json:"deploymentId,string"`
}

// handlerDeployStop stops an ongoing deployment.
func handlerDeployStop(req *app.RequestContext) *shttp.Response {
	dsr := &deployStopRequest{}

	if err := req.Post(dsr); err != nil {
		return shttp.Error(err)
	}

	store := deploy.NewStore()
	depl, err := store.MyDeployment(req.Context(), &deploy.DeploymentsQueryFilters{DeploymentID: dsr.DeploymentID})

	if err != nil {
		return shttp.Error(err)
	}

	if depl == nil {
		return shttp.NotFound()
	}

	if depl.AppID != req.App.ID {
		return shttp.NotAllowed()
	}

	if !depl.ExitCode.Valid {
		if err := store.StopDeployment(req.Context(), depl.ID); err != nil {
			return shttp.Error(err)
		}
	}

	if depl.HasStatusChecks() {
		if err := store.StopStatusChecks(req.Context(), depl.ID); err != nil {
			return shttp.Error(err)
		}
	}

	if depl.GithubRunID.ValueOrZero() != 0 {
		_ = deployservice.Github().StopDeployment(depl.GithubRunID.ValueOrZero())
	}

	return shttp.OK()
}
