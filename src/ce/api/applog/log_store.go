package applog

import (
	"bytes"
	"context"
	"strings"
	"text/template"

	"github.com/stormkit-io/stormkit-io/src/lib/database"
	"github.com/stormkit-io/stormkit-io/src/lib/errors"
	"github.com/stormkit-io/stormkit-io/src/lib/slog"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
)

type statement struct {
	selectLogsForDeployment string
	batchInsert             string
}

var stmt = &statement{
	selectLogsForDeployment: `
		SELECT
			id, timestamp, log_label, log_data
		FROM
			app_logs
		WHERE
			deployment_id = $1 AND
			app_id = $2
			{{ .pagination }}
		ORDER BY
			id {{ .sort }}
		LIMIT
			{{ .limit }}
  `,

	batchInsert: `
		INSERT INTO app_logs (
			app_id,
			host_name,
			timestamp,
			request_id,
			log_label,
			log_data,
			env_id,
			deployment_id
		)
		VALUES
			{{ generateValues 8 (len .) }}
		RETURNING
			id`,
}

// Store is the store to handle log logic.
type Store struct {
	*database.Store

	selectTmpl *template.Template
	batchTmpl  *template.Template
}

// NewStore returns a store instance.
func NewStore() *Store {
	return &Store{
		Store: database.NewStore(),
		batchTmpl: template.Must(
			template.New("batchInsert").
				Funcs(template.FuncMap{"generateValues": utils.GenerateValues}).
				Parse(stmt.batchInsert)),
		selectTmpl: template.Must(
			template.
				New("logs").
				Parse(stmt.selectLogsForDeployment),
		),
	}
}

// InsertLogs inserts application logs to the database.
func (s *Store) InsertLogs(ctx context.Context, logs []*Log) error {
	if len(logs) == 0 {
		return nil
	}

	var qb strings.Builder

	if err := s.batchTmpl.Execute(&qb, logs); err != nil {
		return errors.Wrapf(err, errors.ErrorTypeInternal, "failed to execute batch query template for %d logs", len(logs))
	}

	params := []any{}

	for _, record := range logs {
		params = append(params,
			record.AppID, record.HostName, record.Timestamp,
			record.RequestID, record.Label, record.Data,
			record.EnvironmentID, record.DeploymentID,
		)
	}

	rows, err := s.Query(ctx, qb.String(), params...)

	if err != nil {
		return errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to insert %d logs", len(logs))
	}

	if rows == nil {
		return nil
	}

	defer rows.Close()

	i := 0

	for rows.Next() {
		if err := rows.Scan(&logs[i].ID); err != nil {
			return errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to scan log ID at index %d", i)
		}

		i = i + 1
	}

	return nil
}

// LogQuery represents a log query. It inherits all fields
// of a Log and adds more fields to refine the query.
type LogQuery struct {
	AppID        types.ID
	DeploymentID types.ID
	AfterID      types.ID
	BeforeID     types.ID
	Sort         string // "asc" or "desc"
	Limit        int
}

// Logs return the logs for a given deployment.
func (s *Store) Logs(ctx context.Context, query *LogQuery) ([]*Log, error) {
	var wr bytes.Buffer

	params := []any{
		query.DeploymentID,
		query.AppID,
	}

	sort := "asc" // default sort order

	if query.Sort == "desc" {
		sort = query.Sort
	}

	data := map[string]any{
		"limit":      query.Limit + 1,
		"sort":       sort,
		"pagination": "",
	}

	if query.AfterID > 0 {
		data["pagination"] = " AND id < $3"
		params = append(params, query.AfterID)
	} else if query.BeforeID > 0 {
		data["pagination"] = " AND id > $3"
		params = append(params, query.BeforeID)
	}

	if err := s.selectTmpl.Execute(&wr, data); err != nil {
		return nil, err
	}

	rows, err := s.Query(ctx, wr.String(), params...)

	if err != nil || rows == nil {
		return nil, err
	}

	defer rows.Close()

	logs := []*Log{}

	for rows.Next() {
		l := &Log{
			AppID:        query.AppID,
			DeploymentID: query.DeploymentID,
		}

		if err := rows.Scan(&l.ID, &l.Timestamp, &l.Label, &l.Data); err != nil {
			slog.Errorf("error while scanning logs: %v", err)
			continue
		}

		logs = append(logs, l)
	}

	return logs, nil
}
