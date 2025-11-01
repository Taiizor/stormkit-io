package functiontriggerhandlers

import (
	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/functiontrigger"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
)

func handlerFunctionTriggerDelete(req *app.RequestContext) *shttp.Response {
	tfID := utils.StringToID(req.Query().Get("triggerId"))

	if tfID == 0 {
		return shttp.NotFound()
	}

	store := functiontrigger.NewStore()

	trigger, err := store.ByID(req.Context(), tfID)

	if err != nil {
		return shttp.Error(err)
	}

	if trigger == nil || trigger.EnvID != req.EnvID {
		return shttp.NotFound()
	}

	if err := store.Delete(req.Context(), tfID); err != nil {
		return shttp.Error(err)
	}

	return shttp.OK()
}
