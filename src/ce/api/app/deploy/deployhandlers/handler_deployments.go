package deployhandlers

import (
	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/deploy"
	"github.com/stormkit-io/stormkit-io/src/lib/model"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp/shttperr"
)

const maximumNumberOfDeployments = 20

type deploymentsRequest struct {
	model.Model
	deploy.DeploymentsQueryFilters
}

// Validate validates the deploy list request.
func (dr *deploymentsRequest) Validate() *shttperr.ValidationError {
	err := &shttperr.ValidationError{}

	if dr.From < 0 {
		err.SetError("from", "From cannot be smaller than 0")
	}

	return err.ToError()
}

// handlerDeployments returns the list of deployments. It's possible to filter the deployments.
// For more information check the deploymentsRequest struct.
func handlerDeployments(req *app.RequestContext) *shttp.Response {
	dr := &deploymentsRequest{}

	if err := req.Post(dr); err != nil {
		return shttp.Error(err)
	}

	store := deploy.NewStore()
	filters := &deploy.DeploymentsQueryFilters{
		AppID:          req.App.ID,
		AppDisplayName: req.App.DisplayName,
		Published:      dr.Published,
		Status:         dr.Status,
		EnvID:          dr.EnvID,
		Branch:         dr.Branch,
		From:           dr.From,
		Limit:          maximumNumberOfDeployments,
	}

	var hasNextPage bool
	var depls []*deploy.Deployment
	var err error

	if depls, err = store.Deployments(req.Context(), filters); err != nil {
		return shttp.Error(err)
	}

	if depls == nil {
		depls = []*deploy.Deployment{}
	}

	if len(depls) > maximumNumberOfDeployments {
		depls = depls[:maximumNumberOfDeployments]
		hasNextPage = true
	}

	return &shttp.Response{
		Data: map[string]interface{}{
			"deploys":     depls,
			"hasNextPage": hasNextPage,
		},
	}
}
