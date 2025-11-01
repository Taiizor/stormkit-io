package deploy

import (
	"context"

	"github.com/stormkit-io/stormkit-io/src/ce/api/admin"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/appcache"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/buildconf"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
)

// PublishSettings are the settings for publishing a deployment.
type PublishSettings struct {
	DeploymentID types.ID
	EnvID        types.ID
	Percentage   float64
	NoCacheReset bool
}

// AutoPublish automatically publishes successful deployments if the
// auto publish feature is enabled.
func AutoPublishIfNecessary(ctx context.Context, d *Deployment) error {
	if !d.ShouldPublish || d.ExitCode.ValueOrZero() != 0 || d.Error.ValueOrZero() != "" {
		return nil
	}

	settings := []*PublishSettings{
		{
			EnvID:        d.EnvID,
			DeploymentID: d.ID,
			Percentage:   100,
		},
	}

	return Publish(ctx, settings)
}

// Publish publishes a new deployment.
func Publish(ctx context.Context, settings []*PublishSettings) error {
	if err := NewStore().Publish(ctx, settings...); err != nil {
		return err
	}

	envIDs := map[types.ID]bool{}
	store := buildconf.NewStore()

	for _, s := range settings {
		envID := s.EnvID

		if _, ok := envIDs[envID]; ok {
			continue
		}

		env, err := store.EnvironmentByID(ctx, envID)

		if err != nil {
			return err
		}

		appl, err := app.NewStore().AppByEnvID(ctx, envID)

		if err != nil {
			return err
		}

		if !s.NoCacheReset {
			if err := appcache.Service().Reset(env.ID); err != nil {
				return err
			}
		}

		whs := app.NewStore().OutboundWebhooks(ctx, env.AppID)
		cnf := admin.MustConfig()

		for _, wh := range whs {
			if wh.TriggerOnPublish() {
				wh.Dispatch(app.OutboundWebhookSettings{
					AppID:                  env.AppID,
					DeploymentID:           s.DeploymentID,
					DeploymentStatus:       "success",
					EnvironmentName:        env.Name,
					DeploymentEndpoint:     cnf.PreviewURL(appl.DisplayName, s.DeploymentID.String()),
					DeploymentLogsEndpoint: cnf.DeploymentLogsURL(env.AppID, s.DeploymentID),
				})
			}
		}

		envIDs[envID] = true
	}

	return nil
}
