package jobs

import (
	"context"
	"strings"

	"github.com/stormkit-io/stormkit-io/src/lib/config"
	"github.com/stormkit-io/stormkit-io/src/lib/errors"
	"github.com/stormkit-io/stormkit-io/src/lib/integrations"
	"github.com/stormkit-io/stormkit-io/src/lib/slog"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
)

type KeyContextNumberOfDeploymentsToDelete struct{}

// RemoveDeploymentArtifactsManually removes the artifacts of expired deployments.
// An expired deployment is a deployment that has not been used for more than 30 days.
// Overwrite the numberOfDays to set a custom expiration time.
func RemoveDeploymentArtifactsManually(ctx context.Context, numberOfDays int) ([]string, error) {
	limit, _ := ctx.Value(KeyContextNumberOfDeploymentsToDelete{}).(int)

	if limit <= 0 {
		limit = 100
	}

	if numberOfDays == 0 {
		numberOfDays = 30
	}

	store := NewStore()
	deployments, err := store.DeploymentsOlderThan30Days(ctx, numberOfDays, limit)

	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "failed to fetch old deployments").WithContext("numberOfDays", numberOfDays).WithContext("limit", limit)
	}

	if len(deployments) == 0 {
		return nil, nil
	}

	client := integrations.Client()
	idsToBeMarked := []types.ID{}
	idsToBeMarkedStr := []string{}

	for _, d := range deployments {
		args := integrations.DeleteArtifactsArgs{
			FunctionLocation: d.FunctionLocation.ValueOrZero(),
			APILocation:      d.APILocation.ValueOrZero(),
			StorageLocation:  d.StorageLocation.ValueOrZero(),
		}

		if err := client.DeleteArtifacts(ctx, args); err != nil {
			wrappedErr := errors.Wrap(err, errors.ErrorTypeExternal, "failed to delete artifacts").WithContext("deploymentID", d.ID.String())
			slog.Errorf("error while deleting artifact: %s", wrappedErr.Error())
			continue
		}

		idsToBeMarked = append(idsToBeMarked, d.ID)
		idsToBeMarkedStr = append(idsToBeMarkedStr, d.ID.String())
	}

	if err = store.MarkDeploymentArtifactsDeleted(ctx, idsToBeMarked); err != nil {
		wrappedErr := errors.Wrap(err, errors.ErrorTypeDatabase, "failed to mark artifacts as deleted").WithContext("idsCount", len(idsToBeMarked))
		slog.Errorf("error while marking artifacts deleted: %s", strings.Join(idsToBeMarkedStr, ", "))
		return nil, wrappedErr
	}

	return idsToBeMarkedStr, nil
}

// RemoveDeploymentArtifacts is a job to remove the artifacts of expired deployments.
func RemoveDeploymentArtifacts(ctx context.Context) error {
	idsToBeMarked, err := RemoveDeploymentArtifactsManually(ctx, 30)

	if err != nil {
		return err
	}

	if !config.IsTest() {
		slog.Infof("artifacts deleted: %s", strings.Join(idsToBeMarked, ", "))
	}

	return nil
}
