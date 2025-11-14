package jobs

import (
	"context"

	"github.com/stormkit-io/stormkit-io/src/lib/errors"
	"github.com/stormkit-io/stormkit-io/src/lib/slog"
)

// RemoveStaleEnvironments marks soft deleted environment's deployments soft deleted and
// removes environment if deployments of environment is deleted
func RemoveStaleEnvironments(ctx context.Context) error {
	store := NewStore()

	// if environment is soft deleted, mark deployments belonging to those environments soft
	// deleted as well
	_, err := store.Exec(ctx, stmt.markDeploymentsSoftDeleted)

	if err != nil {
		err = errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to mark deployments of soft deleted environments as deleted")
		slog.Errorf("error while marking deployments soft deleted %v", err)
		return err
	}

	return nil
}
