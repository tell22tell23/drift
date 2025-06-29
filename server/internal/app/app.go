package app

import (
	"database/sql"
	"log"
	"os"

	"github.com/sammanbajracharya/drift/internal/api"
	"github.com/sammanbajracharya/drift/internal/store"
	"github.com/sammanbajracharya/drift/migrations"
)

type Application struct {
	Logger *log.Logger
	DB     *sql.DB

	userHandler *api.UserHandler
}

func NewApplication() (*Application, error) {
	pgDB, err := store.Open()
	if err != nil {
		return nil, err
	}

	migrationDB, err := store.Open()
	if err != nil {
		return nil, err
	}
	defer migrationDB.Close()
	err = store.MigrateFS(migrationDB, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	userStore := store.NewPgUserStore(pgDB)
	accountStore := store.NewPgAccountStore(pgDB)
	sessionStore := store.NewPgSessionStore(pgDB)
	userHandler := api.NewUserHandler(userStore, accountStore, sessionStore, logger)

	return &Application{
		Logger:      logger,
		DB:          pgDB,
		userHandler: userHandler,
	}, nil
}

func (a *Application) UserHandler() *api.UserHandler {
	return a.userHandler
}
