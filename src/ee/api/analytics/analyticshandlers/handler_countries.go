package analyticshandlers

import (
	"fmt"
	"net/http"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ee/api/analytics"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
)

func handlerCountries(req *app.RequestContext) *shttp.Response {
	countries, err := analytics.NewStore().ByCountries(req.Context(), analytics.ByCountriesArgs{
		DomainID: utils.StringToID(req.Query().Get("domainId")),
	})

	if err != nil {
		fmt.Println(err.Error())
		return shttp.Error(err)
	}

	return &shttp.Response{
		Status: http.StatusOK,
		Data:   countries,
	}
}
