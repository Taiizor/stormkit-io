package analyticshandlers

import (
	"net/http"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ee/api/analytics"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
)

func handlerVisitors(req *app.RequestContext) *shttp.Response {
	span := req.Query().Get("ts")

	if span == "" {
		span = analytics.SPAN_24h
	}

	visitors, err := analytics.NewStore().Visitors(req.Context(), analytics.VisitorsArgs{
		Span:       span,
		EnvID:      req.EnvID,
		DomainID:   utils.StringToID(req.Query().Get("domainId")),
		StatusCode: http.StatusOK,
	})

	if err != nil {
		return shttp.Error(err)
	}

	return &shttp.Response{
		Status: http.StatusOK,
		Data:   visitors,
	}
}
