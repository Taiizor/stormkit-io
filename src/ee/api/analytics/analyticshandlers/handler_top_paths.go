package analyticshandlers

import (
	"net/http"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ee/api/analytics"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
)

func handlerTopPaths(req *app.RequestContext) *shttp.Response {
	visitors, err := analytics.NewStore().TopPaths(req.Context(), analytics.TopPathsArgs{
		EnvID:    req.EnvID,
		DomainID: utils.StringToID(req.Query().Get("domainId")),
	})

	if err != nil {
		return shttp.Error(err)
	}

	return &shttp.Response{
		Status: http.StatusOK,
		Data:   visitors,
	}
}
