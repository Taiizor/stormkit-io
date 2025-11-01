package adminhandlers

import (
	"net/http"
	"net/mail"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stormkit-io/stormkit-io/src/ce/api/user"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
)

type ExternalUserRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// handlerExternalUserInvite handles the invitation of a new member to the instance.
// These users will be able to login through a dedicated login screen.
func handlerExternalUserInvite(req *user.RequestContext) *shttp.Response {
	data := &ExternalUserRequest{}

	if err := req.Post(data); err != nil {
		return shttp.Error(err)
	}

	if req.User.HasEmail(data.Email) {
		return &shttp.Response{
			Status: http.StatusBadRequest,
			Data: map[string]string{
				"error": "You cannot invite yourself :)",
			},
		}
	}

	if _, err := mail.ParseAddress(data.Email); err != nil {
		return &shttp.Response{
			Status: http.StatusBadRequest,
			Data: map[string]string{
				"error": "Email is invalid.",
			},
		}
	}

	token, err := req.JWT(jwt.MapClaims{
		"inviterId": req.User.ID.String(),
		"email":     data.Email,
		"role":      data.Role,
	})

	if err != nil {
		return shttp.Error(err)
	}

	return &shttp.Response{
		Status: http.StatusOK,
		Data: map[string]string{
			"token": token,
		},
	}
}
