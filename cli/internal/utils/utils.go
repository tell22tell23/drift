package utils

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	// _libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	peer "github.com/libp2p/go-libp2p/core/peer"
)

func GenerateID() string {
	return "sdkfjhsdkjfh"
}

func GeneratePrivPeerKey() (string, []byte, error) {
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		return "", nil, fmt.Errorf("Error generating key pair: %v", err)
	}

	id, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		return "", nil, fmt.Errorf("Error generating peer ID: %v", err)
	}

	privBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return "", nil, fmt.Errorf("Error marshalling private key: %v", err)
	}

	return id.String(), privBytes, nil
}

func IsRepo() bool {
	dir := ".drift"
	for {
		if _, err := os.Stat(dir); err == nil {
			return true
		}
		if dir == "." {
			break
		}
		dir = filepath.Join("..", dir)
	}
	return false
}

func GetBlob(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error reading file %s: %v", path, err)
	}

	header := fmt.Sprintf("blob %d\x00", len(content))
	blob := append([]byte(header), content...)
	return blob, nil
}

func AddFile(path, repoRoot string) error {
	blob, err := GetBlob(path)
	if err != nil {
		return err
	}
	hashBytes := sha256.Sum256(blob)
	hash := hex.EncodeToString(hashBytes[:])

	relPath, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %v", err)
	}
	entry := fmt.Sprintf("100644 %s %s\n", hash, relPath)

	indexPath := filepath.Join(".drift", "index")
	existingIndex, err := os.ReadFile(indexPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Error reading index file: %v", err)
	}
	if bytes.Contains(existingIndex, []byte(entry)) {
		return nil
	}

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

	f, err := os.OpenFile(indexPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening index file: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("error writing to index file: %v", err)
	}

	return nil
}

func CheckInitialized() error {
	startDir, _ := os.Getwd()
	driftRoot, err := FindDriftRoot(startDir)
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

func FindDriftRoot(startDir string) (string, error) {
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

type TreeEntry struct {
	Mode string
	Type string
	Hash string
	Name string
}

func ParseIndex(data []byte) ([]TreeEntry, error) {
	lines := bytes.Split(data, []byte("\n"))
	entries := []TreeEntry{}

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := bytes.SplitN(line, []byte(" "), 3)
		if len(parts) != 3 {
			continue
		}

		entries = append(entries, TreeEntry{
			Mode: string(parts[0]),
			Type: "blob",
			Hash: string(parts[1]),
			Name: string(parts[2]),
		})
	}

	return entries, nil
}

func BuildTree(entries []TreeEntry, repoRoot string) (string, error) {
	treeMap := map[string][]TreeEntry{}

	for _, e := range entries {
		parts := strings.SplitN(e.Name, string(os.PathSeparator), 2)
		if len(parts) == 1 {
			treeMap["."] = append(treeMap["."], e)
		} else {
			subdir := parts[0]
			e.Name = parts[1]
			treeMap[subdir] = append(treeMap[subdir], e)
		}
	}

	treeEntries := []TreeEntry{}
	for _, e := range treeMap["."] {
		treeEntries = append(treeEntries, e)
	}

	for subdir, subEntries := range treeMap {
		if subdir == "." {
			continue
		}
		subTreeHash, err := BuildTree(subEntries, repoRoot)
		if err != nil {
			return "", err
		}
		treeEntries = append(treeEntries, TreeEntry{
			Mode: "040000",
			Type: "tree",
			Hash: subTreeHash,
			Name: subdir,
		})
	}

	treeData := SerializeTree(treeEntries)
	return WriteObject(repoRoot, "tree", treeData)
}

func SerializeTree(entries []TreeEntry) []byte {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	var buf bytes.Buffer
	for _, e := range entries {
		line := fmt.Sprintf("%s %s %s %s\n", e.Mode, e.Type, e.Hash, e.Name)
		buf.WriteString(line)
	}
	return buf.Bytes()
}

func WriteObject(repoRoot string, objType string, content []byte) (string, error) {
	hdr := []byte(fmt.Sprintf("%s %d\x00", objType, len(content)))
	obj := append(hdr, content...)
	sum := sha256.Sum256(obj)
	id := hex.EncodeToString(sum[:])
	dir := filepath.Join(repoRoot, ".drift", "objects", id[:2])
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("Error creating object directory %s: %v", dir, err)
	}
	path := filepath.Join(dir, id[2:])
	if _, err := os.Stat(path); err == nil {
		return id, nil
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, obj, 0644); err != nil {
		return "", fmt.Errorf("Error writing object file %s: %v", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return "", fmt.Errorf("Error renaming object file %s to %s: %v", tmp, path, err)
	}
	return id, nil
}
