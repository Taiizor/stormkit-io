package jobs

import (
	"bytes"
	"context"
	"text/template"

	"github.com/lib/pq"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/deploy"
	"github.com/stormkit-io/stormkit-io/src/lib/database"
	"github.com/stormkit-io/stormkit-io/src/lib/errors"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
)

// Store represents a store for the deployments and deployment logs.
type Store struct {
	*database.Store
}

// NewStore returns a store instance.
func NewStore() *Store {
	return &Store{database.NewStore()}
}

func (s *Store) RemoveOldLogs(ctx context.Context) error {
	_, err := s.Exec(ctx, stmt.removeOldLogs)
	if err != nil {
		return errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to remove old logs")
	}
	return nil
}

// MarkDeploymentArtifactsDeleted marks the deployment and its artifacts as deleted.
func (s *Store) MarkDeploymentArtifactsDeleted(ctx context.Context, ids []types.ID) error {
	_, err := s.Exec(ctx, stmt.markDeploymentArtifactsDeleted, pq.Array(ids))
	if err != nil {
		return errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to mark %d deployment artifacts as deleted", len(ids))
	}
	return nil
}

// DeploymentsOlderThan30Days returns 100 deployments older than 30 days.
// Returned deployments are not published.
func (s *Store) DeploymentsOlderThan30Days(ctx context.Context, numberOfDays, limit int) ([]*deploy.Deployment, error) {
	var wr bytes.Buffer
	tmpl := template.Must(template.New("old_deployments").Parse(stmt.selectOldOrDeletedDeployments))

	if err := tmpl.Execute(&wr, map[string]any{"days": numberOfDays}); err != nil {
		return nil, errors.Wrapf(err, errors.ErrorTypeInternal, "failed to execute template for deployments older than %d days", numberOfDays)
	}

	rows, err := s.Query(ctx, wr.String(), limit)

	if rows == nil || err != nil {
		if err != nil {
			return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to query deployments older than %d days", numberOfDays)
		}
		return nil, nil
	}

	defer rows.Close()

	var deploys []*deploy.Deployment

	for rows.Next() {
		if deploys == nil {
			deploys = []*deploy.Deployment{}
		}

		d := &deploy.Deployment{}

		err := rows.Scan(
			&d.ID, &d.AppID, &d.StorageLocation, &d.FunctionLocation, &d.APILocation,
		)

		if err != nil {
			return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to scan deployment row")
		}

		deploys = append(deploys, d)
	}

	return deploys, nil
}

func (s *Store) UserIDsWithoutAPIKeys(ctx context.Context) ([]types.ID, error) {
	rows, err := s.Query(ctx, stmt.selectUserIDsWithoutAPIKeys)

	if rows == nil || err != nil {
		return nil, err
	}

	defer rows.Close()

	ids := []types.ID{}

	for rows.Next() {
		var id types.ID

		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}
