package authwall

import (
	"context"

	"github.com/lib/pq"
	"github.com/stormkit-io/stormkit-io/src/lib/database"
	"github.com/stormkit-io/stormkit-io/src/lib/errors"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
)

var stmt = struct {
	createLogin       string
	removeLogins      string
	selectPassword    string
	selectLogins      string
	updateLastLogin   string
	setAuthWallConfig string
}{
	createLogin: `
		INSERT INTO auth_wall
			(env_id, login_email, login_password, created_at)
		VALUES
			($1, $2, $3, NOW())
		 RETURNING
		 	login_id;
	`,
	removeLogins: `
		DELETE FROM auth_wall WHERE env_id = $1 AND login_id = ANY($2);
	`,
	selectPassword: `
		SELECT
			login_id, login_password
		FROM
			auth_wall
		WHERE
			LOWER(login_email) = LOWER($1) AND 
			env_id = $2;
	`,
	selectLogins: `
		SELECT
			login_id, env_id, login_email, login_password, last_login_at, created_at
		FROM
			auth_wall
		WHERE
			env_id = $1
		ORDER BY
			login_id DESC
		LIMIT
			50;
	`,
	updateLastLogin: `
		UPDATE auth_wall SET last_login_at = NOW() AT TIME ZONE 'UTC' WHERE login_id = $1;
	`,
	setAuthWallConfig: `
		UPDATE apps_build_conf SET auth_wall_conf = $1 WHERE env_id = $2;
	`,
}

type store struct {
	*database.Store
}

// Store returns a store instance.
func Store() *store {
	return &store{database.NewStore()}
}

// CreateLogin creates a new login.
func (s *store) CreateLogin(ctx context.Context, aw *AuthWall) error {
	params := []any{
		aw.EnvID,
		aw.LoginEmail,
		utils.EncryptToString(aw.LoginPassword),
	}

	row, err := s.QueryRow(ctx, stmt.createLogin, params...)

	if err != nil {
		return errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to create auth wall login for env_id=%d", aw.EnvID)
	}

	if row == nil {
		return errors.New(errors.ErrorTypeDatabase, "no row returned").WithContext("env_id", aw.EnvID)
	}

	if err := row.Scan(&aw.LoginID); err != nil {
		return errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to scan login ID for env_id=%d", aw.EnvID)
	}

	return nil
}

// RemoveLogins removes a login.
func (s *store) RemoveLogins(ctx context.Context, envID types.ID, loginIDs []types.ID) error {
	_, err := s.Exec(ctx, stmt.removeLogins, envID, pq.Array(loginIDs))
	if err != nil {
		return errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to remove %d auth wall logins for env_id=%d", len(loginIDs), envID)
	}
	return nil
}

// Login validates the login and returns the auth wall struct with the login id.
func (s *store) Login(ctx context.Context, aw *AuthWall) (bool, error) {
	row, err := s.QueryRow(ctx, stmt.selectPassword, aw.LoginEmail, aw.EnvID)

	if err != nil {
		return false, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to query auth wall login for email=%s env_id=%d", aw.LoginEmail, aw.EnvID)
	}

	if row == nil {
		return false, nil
	}

	var password string

	if err := row.Scan(&aw.LoginID, &password); err != nil {
		return false, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to scan auth wall password for env_id=%d", aw.EnvID)
	}

	return utils.DecryptToString(password) == aw.LoginPassword, nil
}

// UpdateLastLogin updates the last login time.
func (s *store) UpdateLastLogin(ctx context.Context, loginID types.ID) error {
	_, err := s.Exec(ctx, stmt.updateLastLogin, loginID)
	if err != nil {
		return errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to update last login for login_id=%d", loginID)
	}
	return nil
}

// Logins returns all logins for an environment.
func (s *store) Logins(ctx context.Context, envID types.ID) ([]AuthWall, error) {
	rows, err := s.Query(ctx, stmt.selectLogins, envID)

	if err != nil {
		return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to query auth wall logins for env_id=%d", envID)
	}

	defer rows.Close()

	var logins []AuthWall

	for rows.Next() {
		aw := AuthWall{}
		err := rows.Scan(
			&aw.LoginID, &aw.EnvID, &aw.LoginEmail, &aw.LoginPassword, &aw.LastLogin, &aw.CreatedAt,
		)

		if err != nil {
			return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to scan auth wall login for env_id=%d", envID)
		}

		logins = append(logins, aw)
	}

	return logins, nil
}

func (s *store) SetAuthWallConfig(ctx context.Context, envID types.ID, cfg *Config) error {
	_, err := s.Exec(ctx, stmt.setAuthWallConfig, cfg, envID)
	if err != nil {
		return errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to set auth wall config for env_id=%d", envID)
	}
	return nil
}

// AuthWallConfig returns the configuration associated with the environment.
func (s *store) AuthWallConfig(ctx context.Context, envID types.ID) (*Config, error) {
	row, err := s.QueryRow(ctx, "SELECT auth_wall_conf FROM apps_build_conf WHERE env_id = $1", envID)

	if err != nil {
		return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to query auth wall config for env_id=%d", envID)
	}

	if row == nil {
		return nil, nil
	}

	cfg := &Config{}

	if err := row.Scan(cfg); err != nil {
		return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to scan auth wall config for env_id=%d", envID)
	}

	return cfg, nil
}
