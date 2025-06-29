package store

import (
	"database/sql"
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Image     *string   `json:"image"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PgUserStore struct {
	db *sql.DB
}

type UserStore interface {
	// CRUD
	CreateUser(user *User) (*User, error)
	GetByID(id string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id string) error

	// Lookups
	GetByEmail(email string) (*User, error)

	// Utility
	UserExistsByEmail(email string) (*User, error)
	ValidateUser(
		email, password string,
		accountStore AccountStore,
	) (*User, error)
}

func NewPgUserStore(db *sql.DB) *PgUserStore {
	return &PgUserStore{db: db}
}

func (pg *PgUserStore) CreateUser(user *User) (*User, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `INSERT INTO users (name, email, image)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at`
	err = tx.
		QueryRow(query, user.Name, user.Email, user.Image).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (pg *PgUserStore) GetByID(id string) (*User, error) {
	user := &User{}
	query := `SELECT id, name, email, image, created_at, updated_at
	FROM users
	WHERE id = $1`
	err := pg.db.
		QueryRow(query, id).
		Scan(&user.ID, &user.Name, &user.Email, &user.Image, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (pg *PgUserStore) GetByEmail(email string) (*User, error) {
	user := &User{}
	query := `SELECT id, name, email, image, created_at, updated_at
	FROM users
	WHERE email = $1`
	err := pg.db.
		QueryRow(query, email).
		Scan(&user.ID, &user.Name, &user.Email, &user.Image, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (pg *PgUserStore) UpdateUser(user *User) error {
	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE users
	SET name = $1, email = $2, image = $3, updated_at = CURRENT_TIMESTAMP
	WHERE id = $4`
	results, err := tx.Exec(query, user.Name, user.Email, user.Image, user.ID)
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

func (pg *PgUserStore) DeleteUser(id string) error {
	query := `DELETE FROM users WHERE id = $1`
	results, err := pg.db.Exec(query, id)
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

func (pg *PgUserStore) UserExistsByEmail(email string) (*User, error) {
	user, err := pg.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	return user, nil
}

func (pg *PgUserStore) ValidateUser(
	email, password string,
	accountStore AccountStore,
) (*User, error) {
	existingUser, err := pg.UserExistsByEmail(email)
	if err != nil {
		return nil, err
	}
	if existingUser == nil {
		return nil, nil
	}

	err = accountStore.ValidatePassword(existingUser.ID, password)
	if err != nil {
		return nil, nil
	}

	return existingUser, nil
}
