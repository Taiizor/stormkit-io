package mailer

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/stormkit-io/stormkit-io/src/lib/database"
	"github.com/stormkit-io/stormkit-io/src/lib/errors"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
)

var stmt = struct {
	selectConfig string
	upsertConfig string
	selectEmails string
	insertEmail  string
}{
	selectConfig: `
		SELECT
			COALESCE(mailer_conf, '{}')
		FROM
			apps_build_conf
		WHERE
			env_id = $1 AND
			deleted_at IS NULL;
	`,

	upsertConfig: `
		UPDATE apps_build_conf SET mailer_conf = $1 WHERE env_id = $2;	
	`,

	selectEmails: `
		SELECT
			email_id, env_id, email_to, email_from,
			email_subject, email_body, created_at
		FROM
			mailer
		WHERE
			env_id = $1
		ORDER BY
			email_id DESC
		LIMIT
			100;
	`,

	insertEmail: `
		INSERT INTO mailer
			(env_id, email_to, email_from, email_subject, email_body)
		VALUES
			($1, $2, $3, $4, $5);
	`,
}

// Store represents a store for volume management.
type store struct {
	*database.Store
}

// Store returns a new store instance.
func Store() *store {
	return &store{
		Store: database.NewStore(),
	}
}

// Config returns the Mailer Configuration for the given environment.
func (s *store) Config(ctx context.Context, envID types.ID) (*Config, error) {
	row, err := s.QueryRow(ctx, stmt.selectConfig, envID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to query mailer config for env_id=%d", envID)
	}

	var data []byte

	if err := row.Scan(&data); err != nil {
		return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to scan mailer config for env_id=%d", envID)
	}

	cnf := &Config{
		EnvID: envID,
	}

	if err := json.Unmarshal(data, cnf); err != nil {
		return nil, errors.Wrapf(err, errors.ErrorTypeInternal, "failed to unmarshal mailer config for env_id=%d", envID)
	}

	if cnf.Host == "" || cnf.Username == "" || cnf.Password == "" {
		return nil, nil
	}

	return cnf, nil
}

// UpsertConfig creates or updates the volumes config.
func (s *store) UpsertConfig(ctx context.Context, cnf *Config) error {
	data, err := json.Marshal(cnf)

	if err != nil {
		return errors.Wrapf(err, errors.ErrorTypeInternal, "failed to marshal mailer config for env_id=%d", cnf.EnvID)
	}

	_, err = s.Exec(ctx, stmt.upsertConfig, data, cnf.EnvID)
	if err != nil {
		return errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to upsert mailer config for env_id=%d", cnf.EnvID)
	}
	return nil
}

// InsertMail inserts a sent email to the database. This is mostly for auditing.
func (s *store) InsertEmail(ctx context.Context, mail Email) error {
	_, err := s.Exec(ctx, stmt.insertEmail, mail.EnvID, mail.To, mail.From, mail.Body, mail.Subject)
	if err != nil {
		return errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to insert email for env_id=%d to=%s", mail.EnvID, mail.To)
	}
	return nil
}

// Emails returns the last sent 100 emails.
func (s *store) Emails(ctx context.Context, envID types.ID) ([]*Email, error) {
	rows, err := s.Query(ctx, stmt.selectEmails, envID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to query emails for env_id=%d", envID)
	}

	defer rows.Close()

	emails := []*Email{}

	for rows.Next() {
		email := &Email{}
		err := rows.Scan(
			&email.ID, &email.EnvID, &email.To, &email.From,
			&email.Subject, &email.Body, &email.SentAt,
		)

		if err != nil {
			return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to scan email row for env_id=%d", envID)
		}

		emails = append(emails, email)
	}

	return emails, nil
}
