package adminhandlers

import (
	"net/http"

	"github.com/stormkit-io/stormkit-io/src/ce/api/user"
	"github.com/stormkit-io/stormkit-io/src/lib/rediscache"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
)

func handlerMiseUpdate(req *user.RequestContext) *shttp.Response {
	services := []string{
		rediscache.ServiceHosting,
		rediscache.ServiceWorkerserver,
	}

	if err := rediscache.SetAll("mise_update", rediscache.StatusSent, services); err != nil {
		return shttp.Error(err)
	}

	if err := rediscache.Broadcast(rediscache.EventMiseUpdate); err != nil {
		if err := rediscache.SetAll("mise_update", rediscache.StatusErr, services); err != nil {
			return shttp.Error(err)
		}

		return shttp.Error(err)
	}

	return &shttp.Response{
		Status: http.StatusOK,
	}
}
