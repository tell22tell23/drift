package store

import (
	"context"
	"database/sql"
	"time"
)

type Repo struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	URL           string    `json:"url"`
	OwnerID       string    `json:"owner_id"`
	AccessUserIDs []string  `json:"access_user_ids"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PgRepoStore struct {
	db *sql.DB
}

type RepoStore interface {
	CreateRepo(repo *Repo) (*Repo, error)

	CheckRepoExistence(repoName string) (bool, error)
}

func NewPgRepoStore(db *sql.DB) *PgRepoStore {
	return &PgRepoStore{db: db}
}

func (pg *PgRepoStore) CreateRepo(repo *Repo) (*Repo, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `INSERT INTO repos (name, description, url, owner_id)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, updated_at`
	ctx := context.Background()
	err = tx.
		QueryRowContext(ctx, query, repo.Name, repo.Description, repo.URL, repo.OwnerID).
		Scan(&repo.ID, &repo.CreatedAt, &repo.UpdatedAt)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (pg *PgRepoStore) CheckRepoExistence(repoName string) (bool, error) {
	ctx := context.Background()
	query := `SELECT EXISTS(SELECT 1 FROM repos WHERE name = $1)`
	var exists bool
	err := pg.db.QueryRowContext(ctx, query, repoName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
