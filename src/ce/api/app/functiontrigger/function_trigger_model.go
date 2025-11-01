package functiontrigger

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
)

type Options struct {
	Method  string        `json:"method"`
	URL     string        `json:"url"`
	Payload []byte        `json:"payload,omitempty"`
	Headers shttp.Headers `json:"headers,omitempty"`
}

func (a Options) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Options) Scan(value any) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)

	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

type FunctionTrigger struct {
	ID        types.ID   `json:"id,string"`
	EnvID     types.ID   `json:"envId,string"`
	Cron      string     `json:"cron"`
	Status    bool       `json:"status"`
	Options   Options    `json:"options,omitempty"`
	NextRunAt utils.Unix `json:"nextRunAt,omitempty"`
	CreatedAt utils.Unix `json:"-"`
	UpdatedAt utils.Unix `json:"-"`
}

// MarshalJSON implements the marshaler interface.
func (t *FunctionTrigger) ToMap() map[string]any {
	return map[string]any{
		"id":        t.ID.String(),
		"envId":     t.EnvID.String(),
		"cron":      t.Cron,
		"status":    t.Status,
		"nextRunAt": t.NextRunAt.Unix(),
		"options": map[string]any{
			"url":     t.Options.URL,
			"headers": t.Options.Headers,
			"method":  t.Options.Method,
			"payload": string(t.Options.Payload),
		},
	}
}

type TriggerLog struct {
	ID        types.ID   `json:"id,string"`
	TriggerID types.ID   `json:"triggerId,string"`
	Request   utils.Map  `json:"request"`
	Response  utils.Map  `json:"response"`
	CreatedAt utils.Unix `json:"createdAt"`
}
