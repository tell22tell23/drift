package core

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sammanbajracharya/drift/internal/store"
	"github.com/sammanbajracharya/drift/internal/utils"
)

type Context struct {
	RepoStore store.RepoStore
}

type Command interface {
	InitRepo() error
	Add(path string) error
}

func (c *Context) InitRepo() error {
	repoPath := ".drift"

	if _, err := os.Stat(".drift"); err == nil {
		return fmt.Errorf("Drift repository already exists")
	}

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return fmt.Errorf("Error creating Drift repository directory")
	}

	subDirs := []string{"objects", "refs/heads", "peers", "sync", "log"}
	errCh := make(chan error, len(subDirs))
	var wg sync.WaitGroup
	for _, dir := range subDirs {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()
			path := filepath.Join(".drift", dir)
			if err := os.MkdirAll(path, 0755); err != nil {
				errCh <- fmt.Errorf("Error creating subdirectory %s", dir)
			}
		}(dir)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	configPath := filepath.Join(repoPath, "config")
	genID := utils.GenerateUUID() // use libp2p peer ID
	config := fmt.Sprintf(`[peer]
		id = %s
		address =
		`, genID)

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("Error writing to config file")
	}

	headPath := filepath.Join(repoPath, "HEAD")
	headContent := "ref: refs/heads/main\n"
	if err := os.WriteFile(headPath, []byte(headContent), 0644); err != nil {
		return fmt.Errorf("Error writing HEAD file")
	}

	return nil
}

// TODO:
// add .driftignore feature
// add garbage collection feature
func (c *Context) Add(path string) error {
	if err := checkInitialized(); err != nil {
		return err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %v", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("file or directory does not exist: %w", err)
	}

	repoRoot, err := findDriftRoot(absPath)
	if err != nil {
		return fmt.Errorf("failed to find Drift repository root: %v", err)
	}

	if info.IsDir() {
		return filepath.WalkDir(absPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() && d.Name() == ".drift" {
				return filepath.SkipDir
			}
			if d.IsDir() {
				return nil
			}
			return addFile(path, repoRoot)
		})
	} else {
		return addFile(absPath, repoRoot)
	}
}

func addFile(path, repoRoot string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Error reading file %s: %v", path, err)
	}

	header := fmt.Sprintf("blob %d\x00", len(content))
	blob := append([]byte(header), content...)

	hashBytes := sha256.Sum256(blob)
	hash := hex.EncodeToString(hashBytes[:])

	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	_, err = w.Write(blob)
	if err != nil {
		return fmt.Errorf("Error compressing file %s: %v", path, err)
	}
	w.Close()

	objectDir := filepath.Join(".drift", "objects", hash[:2])
	objectPath := filepath.Join(objectDir, hash[2:])

	if err := os.MkdirAll(objectDir, 0755); err != nil {
		return fmt.Errorf("Error creating object directory %s: %v", objectDir, err)
	}

	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		err = os.WriteFile(objectPath, compressed.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("Error writing compressed file %s: %v", objectPath, err)
		}
	}

	indexPath := filepath.Join(".drift", "index")
	relPath, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %v", err)
	}
	existing, _ := os.ReadFile(indexPath)
	entry := fmt.Sprintf("100644 %s %s\n", hash, relPath)

	if !bytes.Contains(existing, []byte(entry)) {
		f, err := os.OpenFile(indexPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("error opening index file: %v", err)
		}
		defer f.Close()

		if _, err := f.WriteString(entry); err != nil {
			return fmt.Errorf("error writing to index file: %v", err)
		}
	}

	return nil
}

func checkInitialized() error {
	startDir, _ := os.Getwd()
	driftRoot, err := findDriftRoot(startDir)
	if err != nil {
		return err
	}

	objectsDir := filepath.Join(driftRoot, ".drift", "objects")
	indexPath := filepath.Join(driftRoot, ".drift", "index")

	if _, err := os.Stat(objectsDir); os.IsNotExist(err) {
		return fmt.Errorf("fatal: missing objects directory: %s", objectsDir)
	}
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		err := os.WriteFile(indexPath, []byte{}, 0644)
		if err != nil {
			return fmt.Errorf("error creating index: %v", err)
		}
	}

	return nil
}

func findDriftRoot(startDir string) (string, error) {
	dir := startDir

	for {
		driftPath := filepath.Join(dir, ".drift")
		if stat, err := os.Stat(driftPath); err == nil && stat.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached the root "/"
		}
		dir = parent
	}

	return "", fmt.Errorf(
		"fatal: not a Drift repository (or any of the parent directories): .drift",
	)
}
