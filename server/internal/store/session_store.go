package store

import (
	"database/sql"
	"time"
)

type Session struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
	Token     string    `json:"token"`
	IpAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PgSessionStore struct {
	db *sql.DB
}

type SessionStore interface {
	CreateSession(session *Session) (*Session, error)
	DeleteSessionByToken(token string) error
	GetSessionByToken(token string) (*Session, error)
}

func NewPgSessionStore(db *sql.DB) *PgSessionStore {
	return &PgSessionStore{db: db}
}

func (pg *PgSessionStore) CreateSession(session *Session) (*Session, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `INSERT INTO sessions (id, expires_at, token, ip_address, user_agent, user_id)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING created_at, updated_at`
	err = tx.
		QueryRow(query,
			session.ID,
			session.ExpiresAt,
			session.Token,
			session.IpAddress,
			session.UserAgent,
			session.UserID).
		Scan(&session.CreatedAt, &session.UpdatedAt)

	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return session, nil
}

func (pg *PgSessionStore) DeleteSessionByToken(token string) error {
	query := `DELETE FROM sessions
	WHERE token = $1`
	results, err := pg.db.Exec(query, token)
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

	return nil
}

func (pg *PgSessionStore) GetSessionByToken(token string) (*Session, error) {
	var session Session
	query := `SELECT id, expires_at, token, ip_address, user_agent, user_id, created_at, updated_at
	FROM sessions
	WHERE token = $1`

	err := pg.db.
		QueryRow(query, token).
		Scan(&session.ID, &session.ExpiresAt, &session.Token, &session.IpAddress, &session.UserAgent, &session.UserID, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &session, nil
}
