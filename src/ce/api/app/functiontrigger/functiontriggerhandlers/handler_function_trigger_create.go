package functiontriggerhandlers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/adhocore/gronx"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/functiontrigger"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
)

type FunctionTriggerRequest struct {
	ID      types.ID `json:"id,string"`
	EnvID   types.ID `json:"envId,string"`
	Cron    string   `json:"cron"`
	Status  bool     `json:"status"`
	Options struct {
		Headers shttp.Headers `json:"headers"`
		Method  string        `json:"method"`
		Payload string        `json:"payload"`
		URL     string        `json:"url"`
	} `json:"options"`
}

func validate(data *FunctionTriggerRequest) map[string]string {
	errors := map[string]string{}

	if !gronx.New().IsValid(data.Cron) {
		errors["cron"] = "Invalid cron format"
	}

	// Parse the URL
	parsedURL, err := url.Parse(data.Options.URL)

	if err != nil ||
		parsedURL.Host == "" ||
		parsedURL.Scheme == "" ||
		!strings.EqualFold(parsedURL.Scheme, "http") && !strings.EqualFold(parsedURL.Scheme, "https") {
		errors["url"] = "Invalid URL"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

func handlerFunctionTriggerCreate(req *app.RequestContext) *shttp.Response {
	tf := &FunctionTriggerRequest{}

	if err := req.Post(tf); err != nil {
		return shttp.Error(err)
	}

	if errs := validate(tf); errs != nil {
		return &shttp.Response{
			Status: http.StatusBadRequest,
			Data:   errs,
		}
	}

	record := &functiontrigger.FunctionTrigger{
		Cron:   tf.Cron,
		EnvID:  tf.EnvID,
		Status: tf.Status,
		Options: functiontrigger.Options{
			Method:  tf.Options.Method,
			Headers: tf.Options.Headers,
			URL:     tf.Options.URL,
			Payload: []byte(tf.Options.Payload),
		},
	}

	if err := functiontrigger.NewStore().Insert(req.Context(), record); err != nil {
		return shttp.Error(err)
	}

	return shttp.Created()
}
