package utils

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
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
