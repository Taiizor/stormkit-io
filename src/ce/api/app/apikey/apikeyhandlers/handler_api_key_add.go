package apikeyhandlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app/apikey"
	"github.com/stormkit-io/stormkit-io/src/ce/api/user"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
)

type APIKeyGenerateRequest struct {
	Name   string   `json:"name"`
	Scope  string   `json:"scope"`
	EnvID  types.ID `json:"envId,string"`
	AppID  types.ID `json:"appId,string"`
	TeamID types.ID `json:"teamId,string"`
}

func handlerAPIKeyAdd(req *user.RequestContext) *shttp.Response {
	data := &APIKeyGenerateRequest{}

	if err := req.Post(data); err != nil {
		return shttp.Error(err)
	}

	key := &apikey.Token{
		AppID:  data.AppID,
		EnvID:  data.EnvID,
		TeamID: data.TeamID,
		Name:   strings.TrimSpace(data.Name),
		Scope:  data.Scope,
		Value:  apikey.GenerateTokenValue(),
	}

	if key.Scope == "" {
		key.Scope = apikey.SCOPE_ENV
	}

	if !apikey.IsScopeValid(key.Scope) {
		return &shttp.Response{
			Status: http.StatusBadRequest,
			Data: map[string]string{
				"error": fmt.Sprintf("Invalid scope. Allowes scopes are: %s", strings.Join(apikey.AllowedScopes, ", ")),
			},
		}
	}

	if key.Name == "" {
		return &shttp.Response{
			Status: http.StatusBadRequest,
			Data: map[string]string{
				"error": "Key name is a required field.",
			},
		}
	}

	if err := apikey.NewStore().AddAPIKey(req.Context(), key); err != nil {
		return shttp.Error(err)
	}

	return &shttp.Response{
		Status: http.StatusCreated,
		Data:   key,
	}
}
