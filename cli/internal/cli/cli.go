package cli

import (
	"os"
	"path/filepath"

	"github.com/sammanbajracharya/drift_cli/internal/core"
	"github.com/urfave/cli/v2"
)

type App struct {
	args []string
}

func NewApp() *App {
	return &App{args: os.Args}
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
					ctx := &core.Context{}
					if err := ctx.InitRepo(); err != nil {
						return cli.Exit(err.Error(), 1)
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
					ctx := &core.Context{}
					if err := ctx.Add(path); err != nil {
						return err
					}
					return cli.Exit("Added "+path+" to Drift repository", 0)
				},
			},
			{
				Name:  "status",
				Usage: "Show the status of the Drift repository",
				Action: func(c *cli.Context) error {
					ctx := &core.Context{}
					if err := ctx.Status(); err != nil {
						return err
					}
					return cli.Exit("Drift repository status displayed", 0)
				},
			},
			{
				Name:  "commit",
				Usage: "Commit changes to the Drift repository",
				Action: func(c *cli.Context) error {
					msg := c.Args().First()
					if msg == "" {
						return cli.Exit("Aborting commit due to empty commit message", 1)
					}
					ctx := &core.Context{}
					if err := ctx.Commit(msg); err != nil {
						return err
					}
					return cli.Exit("Changes committed to Drift repository", 0)
				},
			},
			{
				Name:  "config",
				Usage: "Get or set configuration options",
				Action: func(c *cli.Context) error {
					key := c.Args().Get(0)
					value := c.Args().Get(1)
					ctx := &core.Context{}
					if key == "" {
						return cli.Exit("Please specify a configuration key", 1)
					}
					if value == "" {
						val, err := ctx.GetConfig(key)
						if err != nil {
							return err
						}
						return cli.Exit(key+" = "+val, 0)
					} else {
						if err := ctx.SetConfig(key, value); err != nil {
							return err
						}
						return cli.Exit("Set "+key+" to "+value, 0)
					}
				},
			},
		},
	}

	return app.Run(a.args)
}
