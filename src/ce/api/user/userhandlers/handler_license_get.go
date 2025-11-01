package userhandlers

import (
	"net/http"

	"github.com/stormkit-io/stormkit-io/src/ce/api/user"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
)

func handlerLicenseGet(req *user.RequestContext) *shttp.Response {
	license, err := user.NewStore().LicenseByUserID(req.Context(), req.User.ID)

	if err != nil {
		return shttp.Error(err)
	}

	if license == nil {
		return &shttp.Response{
			Status: http.StatusOK,
			Data: map[string]any{
				"license": nil,
			},
		}
	}

	return &shttp.Response{
		Status: http.StatusOK,
		Data: map[string]any{
			"license": map[string]any{
				"seats":      license.Seats,
				"raw":        license.Token(),
				"enterprise": license.IsEnterprise(),
			},
		},
	}
}
