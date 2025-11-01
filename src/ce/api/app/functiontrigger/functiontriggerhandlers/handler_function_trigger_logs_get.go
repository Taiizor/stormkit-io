package functiontriggerhandlers

import (
	"net/http"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/functiontrigger"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
)

func handleTriggerLogsGet(req *app.RequestContext) *shttp.Response {
	triggerID := utils.StringToID(req.URL().Query().Get("triggerId"))

	if triggerID == 0 {
		return shttp.NotFound()
	}

	store := functiontrigger.NewStore()

	tf, err := store.ByID(req.Context(), triggerID)

	if err != nil {
		return shttp.Error(err)
	}

	if tf == nil || tf.EnvID != req.EnvID {
		return shttp.NotFound()
	}

	logs, err := store.Logs(req.Context(), triggerID)

	if err != nil {
		return shttp.Error(err)
	}

	return &shttp.Response{
		Status: http.StatusOK,
		Data: map[string]any{
			"logs": logs,
		},
	}
}
