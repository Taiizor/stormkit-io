package functiontriggerhandlers

import (
	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/functiontrigger"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
)

func handlerFunctionTriggerUpdate(req *app.RequestContext) *shttp.Response {
	tf := &FunctionTriggerRequest{}

	if err := req.Post(tf); err != nil {
		return shttp.Error(err)
	}

	store := functiontrigger.NewStore()
	existing, err := store.ByID(req.Context(), tf.ID)

	if err != nil {
		return shttp.Error(err)
	}

	if existing == nil || existing.EnvID != req.EnvID {
		return shttp.NotFound()
	}

	record := &functiontrigger.FunctionTrigger{
		ID:     tf.ID,
		Cron:   tf.Cron,
		Status: tf.Status,
		Options: functiontrigger.Options{
			Method:  tf.Options.Method,
			Headers: tf.Options.Headers,
			URL:     tf.Options.URL,
			Payload: []byte(tf.Options.Payload),
		},
	}

	if err := functiontrigger.NewStore().Update(req.Context(), record); err != nil {
		return shttp.Error(err)
	}

	return shttp.OK()
}
