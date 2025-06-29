package api

import (
	"context"
	"net/http"
	"time"

	"github.com/sammanbajracharya/drift/internal/utils"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func (uh *UserHandler) SessionAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil || cookie.Value == "" {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{
				"error": "Unauthorized",
			})
			return
		}

		session, err := uh.sessionStore.GetSessionByToken(cookie.Value)
		if err != nil || time.Now().After(session.ExpiresAt) {
			_ = uh.sessionStore.DeleteSessionByToken(cookie.Value)
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Session expired"})
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
