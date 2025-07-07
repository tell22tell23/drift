package utils

import (
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
