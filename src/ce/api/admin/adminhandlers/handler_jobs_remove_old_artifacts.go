package adminhandlers

import (
	"context"
	"net/http"

	"github.com/stormkit-io/stormkit-io/src/ce/api/user"
	jobs "github.com/stormkit-io/stormkit-io/src/ce/workerserver"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
)

func handlerJobsRemoveOldArtifacts(req *user.RequestContext) *shttp.Response {
	ctx := context.WithValue(req.Context(), jobs.KeyContextNumberOfDeploymentsToDelete{}, 50)
	ids, err := jobs.RemoveDeploymentArtifactsManually(ctx, 30)

	if err != nil {
		return &shttp.Response{
			Status: http.StatusInternalServerError,
			Data: map[string]any{
				"error": err.Error(),
			},
		}
	}

	return &shttp.Response{
		Status: http.StatusOK,
		Data: map[string]any{
			"deleted": ids,
		},
	}
}
