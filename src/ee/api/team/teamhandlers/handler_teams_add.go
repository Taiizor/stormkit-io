package teamhandlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gosimple/slug"
	"github.com/stormkit-io/stormkit-io/src/ce/api/user"
	"github.com/stormkit-io/stormkit-io/src/ee/api/team"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
)

type TeamAddRequest struct {
	Name string `json:"name"`
}

func handlerTeamsAdd(req *user.RequestContext) *shttp.Response {
	data := TeamAddRequest{}

	if err := req.Post(&data); err != nil {
		return shttp.Error(err)
	}

	data.Name = strings.TrimSpace(data.Name)

	if data.Name == "" {
		return &shttp.Response{
			Status: http.StatusBadRequest,
			Data: map[string]string{
				"error": "Team name is a required field.",
			},
		}
	}

	store := team.NewStore()
	teams, err := store.Teams(req.Context(), req.User.ID)

	if err != nil {
		return shttp.Error(err)
	}

	if len(teams) >= team.MAX_TEAMS_PER_USER {
		return &shttp.Response{
			Status: http.StatusBadRequest,
			Data: map[string]string{
				"error": fmt.Sprintf("User can have maximum %d teams.", team.MAX_TEAMS_PER_USER),
			},
		}
	}

	newTeam := &team.Team{
		Name: strings.TrimSpace(data.Name),
		Slug: slug.Make(data.Name),
	}

	member := &team.Member{
		UserID: req.User.ID,
		Role:   team.ROLE_OWNER,
		Status: true,
	}

	if err := store.CreateTeam(req.Context(), newTeam, member); err != nil {
		return shttp.Error(err)
	}

	return &shttp.Response{
		Status: http.StatusCreated,
		Data: map[string]string{
			"id":   newTeam.ID.String(),
			"name": newTeam.Name,
			"slug": newTeam.Slug,
		},
	}
}
