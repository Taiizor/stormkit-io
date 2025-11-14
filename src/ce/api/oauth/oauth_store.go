package oauth

import (
	"context"
	"database/sql"
	"strings"

	"github.com/stormkit-io/stormkit-io/src/lib/database"
	"github.com/stormkit-io/stormkit-io/src/lib/errors"
	"github.com/stormkit-io/stormkit-io/src/lib/types"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
	"golang.org/x/oauth2"
)

// Store handles user logic in the database.
type Store struct {
	*database.Store
}

// NewStore returns a store instance.
func NewStore() *Store {
	return &Store{database.NewStore()}
}

// OAuthUser returns the oauth2.Token instance for the given user.
func (s *Store) OAuthUser(ID types.ID, conf *oauth2.Config, provider string) (*User, error) {
	u := &User{
		Token:        &oauth2.Token{},
		ProviderName: provider,
	}

	var personalAccessToken []byte

	row, err := s.QueryRow(context.TODO(), ustmt.selectAuthUser, ID, provider)

	if err != nil {
		return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to query oauth user with ID=%d provider=%s", ID, provider)
	}

	err = row.Scan(
		&u.DisplayName, &u.AccountURI, &u.AccessToken,
		&u.RefreshToken, &u.TokenType, &u.Expiry, &personalAccessToken,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to scan oauth user with ID=%d provider=%s", ID, provider)
	}

	// When there is a personal access token, use that one instead of the access token
	if personalAccessToken != nil {
		decrypted, err := utils.Decrypt(personalAccessToken)

		if err != nil {
			return nil, errors.Wrapf(err, errors.ErrorTypeInternal, "failed to decrypt personal access token for user ID=%d", ID)
		}

		u.AccessToken = string(decrypted)
		return u, nil
	}

	// Refresh the token if necessary
	token, err := conf.TokenSource(context.Background(), u.Token).Token()

	if err != nil {
		return nil, errors.Wrapf(err, errors.ErrorTypeExternal, "failed to refresh oauth token for user ID=%d provider=%s", ID, provider)
	}

	if !strings.EqualFold(u.Token.AccessToken, token.AccessToken) {
		u.Token = token

		if err := s.UpsertToken(ID, u); err != nil {
			return nil, errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to upsert oauth token for user ID=%d", ID)
		}
	}

	return u, nil
}

// UpsertToken updates the access token for the user.
func (s *Store) UpsertToken(ID types.ID, user *User) error {
	t := user.Token
	provider := user.ProviderName

	_, err := s.Exec(
		context.TODO(),
		ustmt.upsertToken,
		// Values
		ID, user.DisplayName, user.AccountURI, provider,
		t.AccessToken, t.RefreshToken, t.TokenType, t.Expiry,
		// On conflict
		user.DisplayName, user.AccountURI,
		t.AccessToken, t.RefreshToken, t.TokenType, t.Expiry,
	)

	if err != nil {
		return errors.Wrapf(err, errors.ErrorTypeDatabase, "failed to upsert oauth token for user ID=%d provider=%s", ID, provider)
	}

	return nil
}
