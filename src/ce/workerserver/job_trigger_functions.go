package jobs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/adhocore/gronx"
	"github.com/hibiken/asynq"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/functiontrigger"
	"github.com/stormkit-io/stormkit-io/src/lib/errors"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/slog"
	"github.com/stormkit-io/stormkit-io/src/lib/tasks"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
)

type FunctionTriggerMessage struct {
	ID        types.ID      `json:"id,string"`
	Payload   []byte        `json:"payload"`
	Headers   shttp.Headers `json:"headers"`
	Method    string        `json:"method"`
	URL       string        `json:"url"`
	NextRunAt utils.Unix    `json:"nextRunAt"`
}

// InvokeDueFunctionTriggers fetches function triggers from the database that are due date. Matching
// function triggers will be prepared and sent to the queue for execution.
func InvokeDueFunctionTriggers(ctx context.Context) error {
	tfs, err := functiontrigger.NewStore().DueTriggers(ctx)

	if err != nil {
		wrappedErr := errors.Wrap(err, errors.ErrorTypeDatabase, "failed to fetch due triggers")
		slog.Errorf("error while selecting due function trigger: %v", wrappedErr)
		return wrappedErr
	}

	messages := []FunctionTriggerMessage{}

	for _, tf := range tfs {
		nextRunAt, err := gronx.NextTickAfter(tf.Cron, time.Now().UTC(), false)

		if err != nil {
			wrappedErr := errors.Wrap(err, errors.ErrorTypeInternal, "failed to calculate next tick").WithMetadata("cron", tf.Cron).WithMetadata("triggerID", tf.ID.String())
			slog.Errorf("error while calculating next tick: %s", wrappedErr.Error())
		}

		messages = append(messages, FunctionTriggerMessage{
			URL:       tf.Options.URL,
			Payload:   tf.Options.Payload,
			Headers:   tf.Options.Headers,
			Method:    tf.Options.Method,
			ID:        tf.ID,
			NextRunAt: utils.UnixFrom(nextRunAt),
		})
	}

	if len(messages) == 0 {
		return nil
	}

	if _, err := tasks.Enqueue(ctx, tasks.TriggerFunctionHttp, messages, nil); err != nil {
		wrappedErr := errors.Wrap(err, errors.ErrorTypeInternal, "failed to enqueue trigger task").WithMetadata("messagesCount", len(messages))
		slog.Errorf("error occurred while enqueuing task %s", wrappedErr.Error())
		return wrappedErr
	}

	return nil
}

// HandleFunctionTrigger handles triggering a function trigger.
func HandleFunctionTrigger(ctx context.Context, t *asynq.Task) error {
	tfs := []FunctionTriggerMessage{}

	if err := json.Unmarshal(t.Payload(), &tfs); err != nil {
		wrappedErr := errors.Wrap(err, errors.ErrorTypeInternal, "failed to unmarshal trigger payload")
		slog.Errorf("HandleTriggerFunction cannot unmarshal payload information: %v", wrappedErr)
		return wrappedErr
	}

	logs := []functiontrigger.TriggerLog{}
	updates := map[types.ID]utils.Unix{}

	for _, tf := range tfs {
		res, err := shttp.NewRequestV2(utils.GetString(tf.Method, shttp.MethodGet), tf.URL).
			Headers(tf.Headers.Make()).
			Payload(tf.Payload).
			Do()

		request := map[string]any{
			"url":     tf.URL,
			"method":  tf.Method,
			"headers": tf.Headers,
			"payload": string(tf.Payload),
		}

		var response map[string]any

		if res != nil {
			response = map[string]any{
				"code": res.StatusCode,
				"body": readBody(res.Response),
			}
		} else if err != nil {
			response = map[string]any{
				"error": err.Error(),
			}
		}

		logs = append(logs, functiontrigger.TriggerLog{
			TriggerID: tf.ID,
			Request:   request,
			Response:  response,
		})

		if err != nil {
			wrappedErr := errors.Wrap(err, errors.ErrorTypeExternal, "trigger function request failed").WithMetadata("url", tf.URL).WithMetadata("triggerID", tf.ID.String())
			slog.Errorf("trigger function request failed %v", wrappedErr)
			continue
		}

		updates[tf.ID] = tf.NextRunAt
	}

	store := functiontrigger.NewStore()

	if err := store.InsertLogs(ctx, logs); err != nil {
		wrappedErr := errors.Wrap(err, errors.ErrorTypeDatabase, "failed to insert trigger logs").WithMetadata("logsCount", len(logs))
		slog.Errorf("error while inserting function trigger logs: %s", wrappedErr.Error())
	}

	if err := store.SetNextRunAt(ctx, updates); err != nil {
		wrappedErr := errors.Wrap(err, errors.ErrorTypeDatabase, "failed to update next run times").WithMetadata("updatesCount", len(updates))
		slog.Errorf("error while inserting function trigger batch updates: %s", wrappedErr.Error())
	}

	return nil
}

func readBody(res *http.Response) string {
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return ""
	}

	defer res.Body.Close()

	return string(body)
}
