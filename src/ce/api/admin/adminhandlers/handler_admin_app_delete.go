package adminhandlers

import (
	"context"
	"strconv"
	"time"

	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/appcache"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/buildconf"
	"github.com/stormkit-io/stormkit-io/src/ce/api/user"
	"github.com/stormkit-io/stormkit-io/src/lib/config"
	"github.com/stormkit-io/stormkit-io/src/lib/discord"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/slog"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
)

func handlerAdminAppDelete(req *user.RequestContext) *shttp.Response {
	appl, err := app.NewStore().AppByID(req.Context(), utils.StringToID(req.Query().Get("appId")))

	if err != nil {
		return shttp.Error(err)
	}

	if appl == nil {
		return shttp.NotFound()
	}

	ustore := user.NewStore()
	usr, err := ustore.UserByID(appl.UserID)

	if err != nil {
		return shttp.Error(err)
	}

	err = ustore.MarkUserAsDeleted(req.Context(), appl.UserID)

	if err != nil {
		return shttp.Error(err)
	}

	envStore := buildconf.NewStore()
	envs, err := envStore.ListEnvironments(context.Background(), appl.ID)

	if err != nil {
		slog.Errorf("failed silently while fetching envs for cache invalidation: %v", err)
	}

	for _, env := range envs {
		if err := appcache.Service().Reset(env.ID); err != nil {
			slog.Errorf("failed silently while invalidating cache: %v", err)
		}
	}

	discord.Notify(config.Get().Reporting.DiscordProductionChannel, discord.Payload{
		Embeds: []discord.PayloadEmbed{
			{
				Title:     "User deleted",
				Timestamp: time.Now().Format(time.RFC3339),
				Fields: []discord.PayloadField{
					{Name: "ID", Value: strconv.FormatInt(int64(appl.UserID), 10)},
					{Name: "Name", Value: usr.FullName()},
					{Name: "DisplayName", Value: usr.Display()},
					{Name: "Email", Value: usr.PrimaryEmail()},
				}},
		},
	})

	return shttp.OK()
}
