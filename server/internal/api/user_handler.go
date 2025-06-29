package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/sammanbajracharya/drift/internal/store"
	"github.com/sammanbajracharya/drift/internal/utils"
)

type CreateUserRequest struct {
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Image     *string `json:"image"`
	Password  string  `json:"password,omitempty"`   // required for credential
	AccountID string  `json:"account_id,omitempty"` // optional, only for OAuth
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserHandler struct {
	userStore    store.UserStore
	accountStore store.AccountStore
	sessionStore store.SessionStore
	logger       *log.Logger
}

func NewUserHandler(
	userStore store.UserStore,
	accountStore store.AccountStore,
	sessionStore store.SessionStore,
	logger *log.Logger,
) *UserHandler {
	return &UserHandler{
		userStore:    userStore,
		accountStore: accountStore,
		sessionStore: sessionStore,
		logger:       logger,
	}
}

// GET /users/me
func (uh *UserHandler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok || userID == "" {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Unauthorized"})
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Unauthorized"})
		return
	}
	sessionToken := cookie.Value

	session, err := uh.sessionStore.GetSessionByToken(sessionToken)
	if err != nil || session == nil || session.ExpiresAt.Before(time.Now()) {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Session expired"})
		return
	}

	user, err := uh.userStore.GetByID(userID)
	if err != nil {
		utils.WriteJSON(
			w,
			http.StatusInternalServerError,
			utils.Envelope{"error": "Error fetching user"},
		)
		return
	}

	utils.WriteJSON(
		w,
		http.StatusOK,
		utils.Envelope{
			"user":       user,
			"expires_at": session.ExpiresAt.Format(time.RFC3339),
		},
	)
}

// GET /users/{id}
func (uh *UserHandler) HandleGetUserById(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.ReadIDParam(r)
	if err != nil {
		uh.logger.Printf("Error Reading ID params: %v\n", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid User ID"})
		return
	}

	searchUser, err := uh.userStore.GetByID(userID)
	if err != nil {
		uh.logger.Printf("Error Fetching User: %v\n", err)
		utils.WriteJSON(
			w,
			http.StatusInternalServerError,
			utils.Envelope{"error": "Internal Server Error"},
		)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"user": searchUser})
}

// GET /users/email/{email}
func (uh *UserHandler) HandleGetUserByEmail(w http.ResponseWriter, r *http.Request) {
	userEmail, err := utils.ReadEmailParam(r)
	if err != nil {
		uh.logger.Printf("Error Reading ID params: %v\n", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid User ID"})
		return
	}

	searchUser, err := uh.userStore.GetByEmail(userEmail)
	if err != nil {
		uh.logger.Printf("Error Fetching User: %v\n", err)
		utils.WriteJSON(
			w,
			http.StatusInternalServerError,
			utils.Envelope{"error": "Internal Server Error"},
		)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"user": searchUser})
}

// GET /auth/{provider}
// GET /auth/{provider}/callback
func (uh *UserHandler) HandleOAuthRedirect(w http.ResponseWriter, r *http.Request) {}
func (uh *UserHandler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {}

// POST /auth/register
func (uh *UserHandler) HandleCredentialCreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		uh.logger.Printf("Error Decoding JSON body: %v\n", err)
		utils.WriteJSON(
			w,
			http.StatusBadRequest,
			utils.Envelope{"error": "Invalid request body"},
		)
		return
	}

	if !utils.IsValidEmail(req.Email) {
		uh.logger.Println("Invalid email format")
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid email format"})
		return
	}

	if len(req.Password) < 6 {
		uh.logger.Println("Password too short")
		utils.WriteJSON(
			w,
			http.StatusBadRequest,
			utils.Envelope{"error": "Password must be at least 6 characters"},
		)
		return
	}

	existingUser, _ := uh.userStore.GetByEmail(req.Email)
	if existingUser != nil {
		uh.logger.Printf("User already exist\n")
		utils.WriteJSON(
			w,
			http.StatusConflict,
			utils.Envelope{"error": "Internal Server Error"},
		)
		return
	}

	newUser := &store.User{
		Name:  req.Name,
		Email: req.Email,
		Image: req.Image,
	}
	createdUser, err := uh.userStore.CreateUser(newUser)
	if err != nil {
		uh.logger.Printf("Error Creating User: %v\n", err)
		utils.WriteJSON(
			w,
			http.StatusInternalServerError,
			utils.Envelope{"error": "Internal Server Error"},
		)
		return
	}

	account := &store.Account{
		AccountId:  utils.GenerateUUID(), // generate uuid for credential based login,
		ProviderID: "credential",
		UserID:     createdUser.ID,
		Password:   req.Password,
	}

	_, err = uh.accountStore.CreateAccount(account)
	if err != nil {
		uh.logger.Printf("Error Account: %v\n", err)
		utils.WriteJSON(
			w,
			http.StatusInternalServerError,
			utils.Envelope{"error": "Internal Server Error"},
		)

		delErr := uh.userStore.DeleteUser(createdUser.ID)
		if delErr != nil {
			uh.logger.Printf(
				"Rollback failed: unable to delete user %s: %v",
				createdUser.ID,
				delErr,
			)
		}
		return
	}

	utils.WriteJSON(
		w,
		http.StatusOK,
		utils.Envelope{"user_id": createdUser.ID},
	)
}

// POST /auth/login
func (uh *UserHandler) HandleCredentialLogin(w http.ResponseWriter, r *http.Request) {
	var user LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		uh.logger.Printf("Error Decoding JSON body: %v\n", err)
		utils.WriteJSON(
			w,
			http.StatusBadRequest,
			utils.Envelope{"error": "Invalid request body"},
		)
		return
	}

	existingUser, err := uh.userStore.ValidateUser(user.Email, user.Password, uh.accountStore)
	if err != nil {
		uh.logger.Printf("Error validating user: %v\n", err)
		utils.WriteJSON(
			w,
			http.StatusInternalServerError,
			utils.Envelope{"error": "Internal Server Error"},
		)
		return
	}

	if existingUser == nil {
		uh.logger.Printf("User does not exist\n")
		utils.WriteJSON(
			w,
			http.StatusUnauthorized,
			utils.Envelope{"error": "Invalid email or password"},
		)
		return
	}

	session := &store.Session{
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Token:     utils.GenerateUUID(),
		IpAddress: utils.GetIPAddr(r),
		UserAgent: r.Header.Get("User-Agent"),
		UserID:    existingUser.ID,
	}

	createdSession, err := uh.sessionStore.CreateSession(session)
	if err != nil {
		uh.logger.Printf("Error creating session: %v\n", err)
		utils.WriteJSON(
			w,
			http.StatusInternalServerError,
			utils.Envelope{"error": "Internal Server Error"},
		)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    createdSession.Token,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		Expires:  createdSession.ExpiresAt,
		SameSite: http.SameSiteStrictMode,
	})

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{
		"message": "Login successful",
		"user_id": existingUser.ID,
	})
}

// POST /auth/logout
func (uh *UserHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		utils.WriteJSON(
			w,
			http.StatusUnauthorized,
			utils.Envelope{"error": "Not logged in"},
		)
		return
	}

	err = uh.sessionStore.DeleteSessionByToken(cookie.Value)
	if errors.Is(err, sql.ErrNoRows) {
		utils.WriteJSON(w, http.StatusOK, utils.Envelope{
			"message": "Already logged out",
		})
	}

	if err != nil {
		uh.logger.Printf("Error deleting session: %v\n", err)
		utils.WriteJSON(
			w,
			http.StatusInternalServerError,
			utils.Envelope{"error": "Internal server error"},
		)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "Logged out successfully"})
}

// PUT /users/me
func (uh *UserHandler) HandleUpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok || userID == "" {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Unauthorized"})
		return
	}

	var user store.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		uh.logger.Printf("Error Decoding JSON body: %v\n", err)
		utils.WriteJSON(
			w,
			http.StatusBadRequest,
			utils.Envelope{"error": "Invalid request body"},
		)
		return
	}

	user.ID = userID

	err := uh.userStore.UpdateUser(&user)
	if err != nil {
		uh.logger.Printf("Error Updating User: %v\n", err)
		utils.WriteJSON(
			w,
			http.StatusInternalServerError,
			utils.Envelope{"error": "Internal Server Error"},
		)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "User updated successfully"})
}

// DELETE /user/{id}
func (uh *UserHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.ReadIDParam(r)
	if err != nil {
		uh.logger.Printf("Error Reading ID: %v\n", err)
		utils.WriteJSON(
			w,
			http.StatusBadRequest,
			utils.Envelope{"error": "Invalid Request Body"},
		)
		return
	}

	err = uh.userStore.DeleteUser(userID)
	if err != nil {
		uh.logger.Printf("Error Deleting User: %v\n", err)
		utils.WriteJSON(
			w,
			http.StatusInternalServerError,
			utils.Envelope{"error": "Internal Server Error"},
		)
		return
	}

	utils.WriteJSON(
		w,
		http.StatusOK,
		utils.Envelope{"message": "User deleted successfully"},
	)
}
