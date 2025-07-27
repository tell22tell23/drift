package cli

import (
	"os"
	"path/filepath"

	"github.com/sammanbajracharya/drift/internal/core"
	"github.com/sammanbajracharya/drift/internal/store"
	"github.com/urfave/cli/v2"
)

type App struct {
	args      []string
	repoStore store.RepoStore
}

func NewApp() (*App, error) {
	pgDB, err := store.Open()
	if err != nil {
		return nil, err
	}

	repoStore := store.NewPgRepoStore(pgDB)
	args := os.Args

	return &App{args: args, repoStore: repoStore}, nil
}

func (a *App) Run() error {
	app := &cli.App{
		Name:  "drift",
		Usage: "P2P repository for managing and sharing code",
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "Initialize a new Drift repository",
				Action: func(c *cli.Context) error {
					ctx := &core.Context{RepoStore: a.repoStore}
					if err := ctx.InitRepo(); err != nil {
						return cli.Exit(
							err.Error(), 1,
						)
					}
					dir, err := os.Getwd()
					if err != nil {
						return cli.Exit("Error getting current user", 1)
					}
					return cli.Exit(
						"Initialized empty Drift repository in "+filepath.Join(dir, ".drift"),
						0,
					)
				},
			},
			{
				Name:  "add",
				Usage: "Add a file or folder to the Drift repository",
				Action: func(c *cli.Context) error {
					path := c.Args().First()
					if path == "" {
						return cli.Exit("Please specify a file or folder to add", 1)
					}
					ctx := &core.Context{RepoStore: a.repoStore}
					if err := ctx.Add(path); err != nil {
						return err
					}
					return cli.Exit("Added "+path+" to Drift repository", 0)
				},
			},
		},
	}

	return app.Run(a.args)
}
