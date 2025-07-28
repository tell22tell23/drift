package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	Status() error
}

type IndexEntry struct {
	Hash string
	Seen bool
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
	if err := utils.CheckInitialized(); err != nil {
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

	repoRoot, err := utils.FindDriftRoot(absPath)
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
			return utils.AddFile(path, repoRoot)
		})
	} else {
		return utils.AddFile(absPath, repoRoot)
	}
}

func (c *Context) Status() error {
	if err := utils.CheckInitialized(); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	repoRoot, err := utils.FindDriftRoot(cwd)
	if err != nil {
		return fmt.Errorf("failed to find Drift repository root: %v", err)
	}

	headFile, err := os.ReadFile(filepath.Join(repoRoot, ".drift", "HEAD"))
	if err != nil {
		return fmt.Errorf("failed to read HEAD file: %v", err)
	}
	parts := strings.Split(strings.TrimSpace(string(headFile)), "/")
	branchName := parts[len(parts)-1]

	indexPath := filepath.Join(repoRoot, ".drift", "index")
	indexMap := map[string]*IndexEntry{}
	if data, err := os.ReadFile(indexPath); err == nil {
		lines := bytes.Split(data, []byte("\n"))
		for _, line := range lines {
			if len(line) == 0 {
				continue
			}
			parts := bytes.SplitN(line, []byte(" "), 3)
			if len(parts) < 3 {
				continue
			}
			indexMap[string(parts[2])] = &IndexEntry{Hash: string(parts[1]), Seen: false}
		}
	}

	modifiedFiles := []string{}
	untrackedFiles := []string{}
	deletedFiles := []string{}

	err = filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == ".drift" {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}
		blob, err := utils.GetBlob(path)
		if err != nil {
			return fmt.Errorf("failed to get blob for %s: %v", path, err)
		}
		hashBytes := sha256.Sum256(blob)
		hash := hex.EncodeToString(hashBytes[:])

		if entry, ok := indexMap[relPath]; ok {
			entry.Seen = true
			if entry.Hash != hash {
				modifiedFiles = append(modifiedFiles, relPath)
			}
		} else {
			untrackedFiles = append(untrackedFiles, relPath)
		}

		return nil
	})

	for relPath, entry := range indexMap {
		if !entry.Seen {
			deletedFiles = append(deletedFiles, relPath)
		}
	}

	fmt.Printf("On branch %s\n", branchName)
	fmt.Println("Your branch is up to date with 'origin/" + branchName + "'.")
	fmt.Println()

	if len(modifiedFiles) > 0 {
		fmt.Println("Changes not staged for commit:")
		fmt.Println(`  (use "drift add <file>..." to update what will be committed)`)
		fmt.Println(`  (use "drift restore <file>..." to discard changes in working directory)`)
		for _, file := range modifiedFiles {
			fmt.Printf("        \033[34mmodified:   %s\033[0m\n", file)
		}
		fmt.Println()
	}

	if len(untrackedFiles) > 0 {
		fmt.Println("Untracked files:")
		fmt.Println(`  (use "drift add <file>..." to include in what will be committed)`)
		for _, file := range untrackedFiles {
			fmt.Printf("        \033[31m%s\033[0m\n", file)
		}
		fmt.Println()
	}

	if len(deletedFiles) > 0 {
		fmt.Println("Deleted files:")
		fmt.Println(`  (use "drift restore <file>..." to restore or "drift rm" to remove)`)
		for _, file := range deletedFiles {
			fmt.Printf("        \033[31mdeleted:   %s\033[0m\n", file)
		}
		fmt.Println()
	}

	if len(modifiedFiles) == 0 && len(untrackedFiles) == 0 {
		fmt.Println("nothing to commit, working tree clean")
	} else {
		fmt.Println("no changes added to commit (use \"drift add\" and/or \"drift commit -a\")")
	}

	return nil
}
