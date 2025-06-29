package store

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func Open() (*sql.DB, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return nil, errors.New("Environment variable not set")
	}
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("Failed to open databse: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to open databse: %v", err)
	}

	log.Printf("Connected to database\n")
	return db, nil
}

func MigrateFS(db *sql.DB, migrationFS fs.FS, dir string) error {
	goose.SetBaseFS(migrationFS)
	defer func() {
		goose.SetBaseFS(nil)
	}()
	return Migrate(db, dir)
}

func Migrate(db *sql.DB, dir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migration error: %w", err)
	}

	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("goose up: %v", err)
	}

	return nil
}
