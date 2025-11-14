package database

import (
	"database/sql"
	"strings"
	"time"

	"github.com/stormkit-io/stormkit-io/src/lib/errors"
	"github.com/stormkit-io/stormkit-io/src/lib/slog"
	"golang.org/x/net/context"
)

// SystemRow represents a new system row.
type SystemRow struct {
	MasterInstance   string
	MasterLastAccess time.Time
}

// Store represents a generic store
type Store struct {
	Conn *sql.DB
}

// NewStore returns a new store instance.
func NewStore() *Store {
	return &Store{
		Conn: Connection(),
	}
}

// Query is a wrapper around the sql.Stmt.QueryContext method.
// It prepares and executes the query
func (s *Store) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	stmt, err := s.Prepare(ctx, query)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to prepare query")
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to execute query")
	}

	return rows, nil
}

// QueryRow is a wrapper around the sql.Stmt.QueryRowContext method.
// It prepares and executes the query
func (s *Store) QueryRow(ctx context.Context, query string, args ...any) (*sql.Row, error) {
	stmt, err := s.Prepare(ctx, query)

	if err != nil {
		return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to prepare query")
	}

	defer stmt.Close()

	return stmt.QueryRowContext(ctx, args...), nil
}

// Exec is a wrapper around the sql.Stmt.ExecContext method.
// It prepares and executes the query
func (s *Store) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	stmt, err := s.Prepare(ctx, query)

	if err != nil {
		return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to prepare query")
	}

	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to execute query")
	}

	return result, nil
}

// Prepare prepares a new statement with context and returns it.
func (s *Store) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	stmt, err := s.Conn.PrepareContext(ctx, query)

	if err != nil {
		isContextCanceled := strings.EqualFold(strings.TrimSpace(err.Error()), "context canceled")

		if isContextCanceled {
			return nil, err
		}

		slog.Errorf("error while preparing query=%s, err=%v", query, err)
		return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to prepare statement: query=%s", query)
	}

	if stmt == nil {
		return nil, errors.Wrap(errors.ErrDatabaseConnection, errors.ErrorTypeDatabase, "prepared statement is nil")
	}

	return stmt, nil
}
