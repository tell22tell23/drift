package utils

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type Envelope map[string]interface{}

func WriteJSON(w http.ResponseWriter, status int, data Envelope) error {
	js, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	js = append(js, '\n')
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func ReadIDParam(r *http.Request) (string, error) {
	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		return "", errors.New("Invalid ID Parameter")
	}

	return idParam, nil
}

func ReadEmailParam(r *http.Request) (string, error) {
	email := chi.URLParam(r, "email")
	if email == "" {
		return "", errors.New("Invalid Email Parameter")
	}
	return email, nil
}

func GenerateUUID() string {
	return uuid.New().String()
}

func GetIPAddr(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func CheckRepoExistence(addr string) (bool, error) {
	// addr should be in `dft@<repo-name>.drift` format
	parts := strings.Split(addr, "@")
	if parts[0] != "dft" || len(parts) != 2 {
		return false, errors.New("Invalid address format, expected dft@<repo-name>.drift")
	}
	parts = strings.Split(parts[1], ".")
	if parts[1] != "drift" || len(parts) != 2 {
		return false, errors.New("Invalid address format, expected dft@<repo-name>.drift")
	}

	fmt.Printf("repo name: %s\n", parts[0])
	// TODO: check repo name existence in the database

	return true, nil
}

func UpdateConfig(configPath, addr string) error {
	return nil
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
