package store

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func checkPassword(hash []byte, password string) error {
	return bcrypt.CompareHashAndPassword(hash, []byte(password))
}

type Account struct {
	ID                    string    `json:"id"`
	AccountId             string    `json:"account_id"`
	ProviderID            string    `json:"provider_id"` // e.g., "google", "credential"
	UserID                string    `json:"user_id"`
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	IDToken               string    `json:"id_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	Password              string    `json:"password,omitempty"` // temporary field for password input
	PasswordHash          []byte    `json:"-"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type PgAccountStore struct {
	db *sql.DB
}

type AccountStore interface {
	// Create
	CreateAccount(account *Account) (*Account, error)

	// Read (For credential-based accounts)
	GetByID(id string) (*Account, error)
	GetByUserID(userID, providerID string) (*Account, error)
	GetByEmail(email string) (*Account, error)

	// Read (For OAuth accounts)
	GetByAccountIDAndProviderID(accountID, providerID string) (*Account, error)
	GetByAccessToken(accessToken string) (*Account, error)

	// OAuth Token Flow
	ValidateAccessToken(accountID, providerID, accessToken string) error
	ValidateAndRefreshAccessToken(accountID, providerID, refreshToken string) (*Account, error)

	// Password Management (for credential-based accounts)
	ChangePassword(userID, password string) error
	ValidatePassword(uesrID, password string) error
}

func NewPgAccountStore(db *sql.DB) *PgAccountStore {
	return &PgAccountStore{db: db}
}

func (pg *PgAccountStore) CreateAccount(account *Account) (*Account, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if account.Password != "" {
		hash, err := hashPassword(account.Password)
		if err != nil {
			return nil, err
		}
		account.PasswordHash = hash
	}

	query := `INSERT INTO accounts (account_id, provider_id, user_id, access_token, refresh_token, id_token, access_token_expires_at, refresh_token_expires_at, password_hash)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id, created_at, updated_at`
	err = tx.
		QueryRow(
			query,
			account.AccountId,
			account.ProviderID,
			account.UserID,
			account.AccessToken,
			account.RefreshToken,
			account.IDToken,
			account.AccessTokenExpiresAt,
			account.RefreshTokenExpiresAt,
			account.PasswordHash,
		).
		Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (pg *PgAccountStore) GetByID(id string) (*Account, error) {
	account := &Account{}
	query := `SELECT id, account_id, provider_id, user_id, access_token, refresh_token, id_token, access_token_expires_at, refresh_token_expires_at, password_hash, created_at, updated_at
	FROM accounts
	WHERE id = $1`
	err := pg.db.
		QueryRow(query, id).
		Scan(&account.ID, &account.AccountId, &account.ProviderID, &account.UserID, &account.AccessToken, &account.RefreshToken, &account.IDToken, &account.AccessTokenExpiresAt, &account.RefreshTokenExpiresAt, &account.PasswordHash, &account.CreatedAt, &account.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

func (pg *PgAccountStore) GetByUserID(userID, providerID string) (*Account, error) {
	account := &Account{}
	query := `SELECT id, account_id, provider_id, user_id, access_token, refresh_token, id_token, access_token_expires_at, refresh_token_expires_at, password_hash, created_at, updated_at
	FROM accounts
	WHERE user_id = $1 AND provider_id = $2`
	err := pg.db.
		QueryRow(query, userID, providerID).
		Scan(&account.ID, &account.AccountId, &account.ProviderID, &account.UserID, &account.AccessToken, &account.RefreshToken, &account.IDToken, &account.AccessTokenExpiresAt, &account.RefreshTokenExpiresAt, &account.PasswordHash, &account.CreatedAt, &account.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

func (pg *PgAccountStore) GetByEmail(email string) (*Account, error) {
	account := &Account{}
	query := `SELECT id, account_id, provider_id, user_id, access_token, refresh_token, id_token, access_token_expires_at, refresh_token_expires_at, password_hash, created_at, updated_at
	FROM accounts
	WHERE email = $1`
	err := pg.db.
		QueryRow(query, email).
		Scan(
			&account.ID,
			&account.AccountId,
			&account.ProviderID,
			&account.UserID,
			&account.AccessToken,
			&account.RefreshToken,
			&account.IDToken,
			&account.AccessTokenExpiresAt,
			&account.RefreshTokenExpiresAt,
			&account.PasswordHash,
			&account.CreatedAt,
			&account.UpdatedAt,
		)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

func (pg *PgAccountStore) GetByAccountIDAndProviderID(
	accountID, providerID string,
) (*Account, error) {
	account := &Account{}
	query := `SELECT id, account_id, provider_id, user_id, access_token, refresh_token, id_token, access_token_expires_at, refresh_token_expires_at, password_hash, created_at, updated_at
	FROM accounts
	WHERE account_id = $1 AND provider_id = $2`
	err := pg.db.
		QueryRow(query, accountID, providerID).
		Scan(&account.ID, &account.AccountId, &account.ProviderID, &account.UserID, &account.AccessToken, &account.RefreshToken, &account.IDToken, &account.AccessTokenExpiresAt, &account.RefreshTokenExpiresAt, &account.PasswordHash, &account.CreatedAt, &account.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

func (pg *PgAccountStore) GetByAccessToken(accessToken string) (*Account, error) {
	account := &Account{}
	query := `
	SELECT id, account_id, provider_id, user_id, access_token, refresh_token, id_token,
	       access_token_expires_at, refresh_token_expires_at, password_hash, created_at, updated_at
	FROM accounts
	WHERE access_token = $1
	`

	err := pg.db.QueryRow(query, accessToken).
		Scan(&account.ID, &account.AccountId, &account.ProviderID, &account.UserID,
			&account.AccessToken, &account.RefreshToken, &account.IDToken,
			&account.AccessTokenExpiresAt, &account.RefreshTokenExpiresAt,
			&account.PasswordHash, &account.CreatedAt, &account.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Check expiry here or in caller?
	if time.Now().After(account.AccessTokenExpiresAt) {
		return nil, errors.New("access token expired")
	}

	return account, nil
}

func (pg *PgAccountStore) ValidateAccessToken(accountID, providerID, accessToken string) error {
	account, err := pg.GetByAccountIDAndProviderID(accountID, providerID)
	if err != nil {
		return err
	}

	if account == nil || account.AccessToken != accessToken ||
		time.Now().After(account.AccessTokenExpiresAt) {
		return errors.New("invalid or expired access token")
	}

	return nil
}

func (pg *PgAccountStore) ValidateAndRefreshAccessToken(
	accountID, providerID, refreshToken string,
) (*Account, error) {
	// do we generate new accesstoken our self or ask google for it?
	account, err := pg.GetByAccountIDAndProviderID(accountID, providerID)
	if err != nil {
		return nil, err
	}

	if account == nil {
		return nil, sql.ErrNoRows
	}

	if account.RefreshToken != refreshToken {
		return nil, errors.New("invalid refresh token")
	}

	// 3. Call Google OAuth token endpoint to get new access token
	//    (This involves HTTP request to Google with your client credentials and refresh token)
	//    For example (pseudo-code):
	/*
	   tokenResponse, err := googleOAuthRefreshTokenRequest(refreshToken)
	   if err != nil {
	       return nil, err
	   }
	*/

	// 4. Update account with new access_token, refresh_token (if any), expiry times
	// 5. Save updated account info in DB
	// 6. Return updated account

	return nil, errors.New("not implemented")
}

func (pg *PgAccountStore) ChangePassword(userID, password string) error {
	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	hash, err := hashPassword(password)
	if err != nil {
		return err
	}

	query := `UPDATE accounts
	SET password_hash = $1, updated_at = NOW()
	WHERE user_id = $2`
	results, err := tx.Exec(query, hash, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}

func (pg *PgAccountStore) ValidatePassword(userID, password string) error {
	account, err := pg.GetByUserID(userID, "credential")
	if err != nil {
		return err
	}

	if account == nil {
		return sql.ErrNoRows
	}

	if err = checkPassword(account.PasswordHash, password); err != nil {
		return errors.New("invalid password")
	}

	return nil
}
