package teamhandlers

import (
	"net/http"

	"github.com/stormkit-io/stormkit-io/src/ce/api/user"
	"github.com/stormkit-io/stormkit-io/src/ee/api/team"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
)

func handlerTeamMemberRemove(req *user.RequestContext) *shttp.Response {
	teamID := utils.StringToID(req.Query().Get("teamId"))
	memberID := utils.StringToID(req.Query().Get("memberId"))
	store := team.NewStore()

	myTeam, err := store.Team(req.Context(), teamID, req.User.ID)

	if err != nil {
		return shttp.Error(err)

	}

	member, err := store.TeamMember(req.Context(), memberID)

	if err != nil {
		return shttp.Error(err)
	}

	if member.Role == team.ROLE_OWNER {
		return &shttp.Response{
			Status: http.StatusBadRequest,
			Data: map[string]string{
				"error": "You are the only owner in this team. Delete the team instead.",
			},
		}
	}

	if !team.HasWriteAccess(myTeam.CurrentUserRole) {
		return shttp.NotAllowed()
	}

	if err := store.RemoveTeamMember(req.Context(), teamID, memberID); err != nil {
		return shttp.Error(err)
	}

	return shttp.OK()
}
