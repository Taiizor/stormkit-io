package deployhandlers

import (
	"net/http"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/deploy"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/deployservice"
	"github.com/stormkit-io/stormkit-io/src/ce/api/oauth"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
)

type RestartRequestData struct {
	DeployID types.ID `json:"deploymentId,string"`
}

// handlerDeployStart starts the deployment process for the given app.
// This handler is triggered when the user submits a deploy request
// through the user interface.
func handlerDeployRestart(req *app.RequestContext) *shttp.Response {
	data := &RestartRequestData{}

	if err := req.Post(data); err != nil {
		return shttp.Error(err)
	}

	if data.DeployID == 0 {
		return &shttp.Response{
			Status: http.StatusBadRequest,
			Data: map[string]string{
				"error": "Deployment ID is a required field.",
			},
		}
	}

	store := deploy.NewStore()
	deployment, err := store.DeploymentByID(req.Context(), data.DeployID)

	if err != nil {
		return shttp.Error(err)
	}

	if deployment == nil {
		return shttp.NotFound()
	}

	if deployment.Status() != "failed" {
		return &shttp.Response{
			Status: http.StatusBadRequest,
			Data: map[string]string{
				"error": "Only failed deployments can be restarted.",
			},
		}
	}

	if err := store.Restart(req.Context(), deployment); err != nil {
		return shttp.Error(err)
	}

	deployment.CheckoutRepo = req.App.Repo

	if err = deployservice.New().Deploy(req.Context(), req.App, deployment); err != nil {
		if err == oauth.ErrRepoNotFound || err == oauth.ErrCredsInvalidPermissions {
			return &shttp.Response{
				Status: http.StatusNotFound,
				Data: map[string]string{
					"error": "Repository is not found or is inaccessible.",
				},
			}
		}

		return shttp.Error(err)
	}

	return &shttp.Response{
		Data:  deployment,
		Error: err,
	}
}
