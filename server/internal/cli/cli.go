package cli

import (
	"os"
	"path/filepath"

	"github.com/sammanbajracharya/drift/internal/core"
	"github.com/sammanbajracharya/drift/internal/store"
	"github.com/sammanbajracharya/drift/internal/utils"
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
	coreCtx := &core.Context{
		RepoStore: a.repoStore,
	}

	app := &cli.App{
		Name:  "drift",
		Usage: "P2P repository for managing and sharing code",
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "Initialize a new Drift repository",
				Action: func(c *cli.Context) error {
					if err := core.InitRepo(); err != nil {
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
					if err := core.Add(path); err != nil {
						return err
					}
					return cli.Exit("Added "+path+" to Drift repository", 0)
				},
			},
			{
				Name:  "host",
				Usage: "Host a Drift repository",
				Action: func(c *cli.Context) error {
					init := utils.IsRepo()
					if !init {
						return cli.Exit(
							"Please run 'drift init' to create a new Drift repository",
							1,
						)
					}
					cmd := c.Args().Get(0)
					addr := c.Args().Get(1)

					switch cmd {
					case "init":
						if addr == "" {
							return cli.Exit("Please specify an address to host", 1)
						}
						if err := coreCtx.HostInit(addr); err != nil {
						}

					}

					return nil
				},
			},
			{
				Name:  "connect",
				Usage: "Connect to an existing Drift repository",
				Action: func(c *cli.Context) error {
					// Command can be either
					// Just `drift connect`, which will show you connected peers
					// or `drift connect <cmd> <addr>`
					// This to connect to a specific peer or perform an action
					init := utils.IsRepo()
					if !init {
						return cli.Exit(
							"Please run 'drift init' to create a new Drift repository",
							1,
						)
					}
					cmd := c.Args().Get(0)
					addr := c.Args().Get(1)

					switch cmd {
					case "":
						// No subcommand: list peer
						if err := coreCtx.ConnectList(); err != nil {
							return cli.Exit("Error listing connected peers: "+err.Error(), 1)
						}
						return nil
					case "add":
						if addr == "" {
							return cli.Exit("Please specify an address to add", 1)
						}
						if err := coreCtx.Connect(addr); err != nil {
							return cli.Exit(err.Error(), 1)
						}
						// a.Logger.Printf("Connected to peer %s\n", addr)
						return cli.Exit("Connected to peer "+addr, 0)
					case "remove":
						if addr == "" {
							return cli.Exit("Please specify an address to remove", 1)
						}
						if err := coreCtx.ConnectRemove(addr); err != nil {
							return cli.Exit("Error removing peer: "+err.Error(), 1)
						}
						// a.Logger.Printf("Disconnected from peer %s\n", addr)
						return nil
					default:
						return cli.Exit("Unknown connect subcommand: "+cmd, 1)

					}
				},
			},
		},
	}

	return app.Run(a.args)
}
