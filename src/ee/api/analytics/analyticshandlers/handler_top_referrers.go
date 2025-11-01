package analyticshandlers

import (
	"net/http"
	"strings"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ee/api/analytics"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
)

func handlerTopReferrers(req *app.RequestContext) *shttp.Response {
	query := req.Query()

	visitors, err := analytics.NewStore().TopReferrers(req.Context(), analytics.TopReferrersArgs{
		EnvID:       req.EnvID,
		RequestPath: strings.ToLower(strings.TrimSpace(query.Get("requestPath"))),
		DomainID:    utils.StringToID(req.Query().Get("domainId")),
	})

	if err != nil {
		return shttp.Error(err)
	}

	return &shttp.Response{
		Status: http.StatusOK,
		Data:   visitors,
	}
}
