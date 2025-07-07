package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/sammanbajracharya/drift/internal/store"
	"github.com/sammanbajracharya/drift/internal/utils"
)

type Context struct {
	RepoStore store.RepoStore
}

type Command interface {
	InitRepo() error
	Connect(addr string) error
	ConnectList() error
	ConnectAdd(peer string) error
	ConnectRemove(peer string) error
}

func InitRepo() error {
	root := "."
	repoPath := ".drift"

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip this entry
		}

		if d.IsDir() && d.Name() == repoPath {
			return fmt.Errorf("Drift Repository already exists")
		}
		return nil
	})

	if err != nil {
		return err
	}

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return fmt.Errorf("Error creating Drift repository directory")
	}

	subDirs := []string{"objects", "refs/heads", "peers", "sync", "log"}
	for _, dir := range subDirs {
		path := filepath.Join(".drift", dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("Error creating subdirectory %s", dir)
		}
	}

	configPath := filepath.Join(repoPath, "config")
	configFile, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("Error creating config file")
	}
	defer configFile.Close()

	genID := utils.GenerateUUID() // use libp2p peer ID
	config := fmt.Sprintf(`[peer]
		id = %s
		address =
		`, genID)

	if _, err := configFile.WriteString(config); err != nil {
		return fmt.Errorf("Error writing to config file")
	}

	return nil
}

func (ctx *Context) ConnectList() error {
	// Here you would implement the logic to list connected peers.
	return nil
}

func (ctx *Context) Connect(addr string) error {
	// 1. Check if the addr associated drift repository exists
	repoExists, err := ctx.RepoStore.CheckRepoExistence(addr)
	if err != nil {
		return fmt.Errorf("%v", err.Error())
	}
	if !repoExists {
		return fmt.Errorf("no Drift repository found at %s", addr)
	}
	// 2. set the address in the config file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}
	configPath := filepath.Join(homeDir, ".drift", "config.yaml")
	err = utils.UpdateConfig(configPath, addr) // just a stub function
	if err != nil {
		return fmt.Errorf("failed to update config file")
	}
	// 3. Connect to the peer
	h, _ := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/4001"),
	)
	h.SetStreamHandler("/drift/sync/1.0.0", func(s network.Stream) {
		fmt.Printf("New incomming stream")
		go handleSync(s)
	})
	// 4. Create & Update ~/.drift/remotes.yaml, ~/.drift/config.yaml
	// 5. Add content of .drift/peers
	// complete this

	return nil
}

func (ctx *Context) ConnectRemove(addr string) error {
	// Here you would implement the logic to connect to a Drift repository.
	// For now, we just return nil to indicate success.
	return nil
}

func Add(path string) error {
	// Here you would implement the logic to add a file or directory to the Drift repository.
	// For now, we just return nil to indicate success.
	return nil
}

func handleSync(s network.Stream) {
	// Handle the incoming stream for synchronization.
	// This is a stub function for now.
	fmt.Printf("Handling sync stream from %s\n", s.Conn().RemotePeer())
	defer s.Close()

	// You can read from the stream and write to it as needed.
	// For example:
	// buf := make([]byte, 1024)
	// n, err := s.Read(buf)
	// if err != nil {
	//     fmt.Printf("Error reading from stream: %v\n", err)
	//     return
	// }
	// fmt.Printf("Received data: %s\n", string(buf[:n]))
}
