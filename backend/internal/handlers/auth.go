package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"consensus/internal/auth"
	"consensus/internal/models"
)

// minPasswordLength is intentionally basic — this is a placeholder auth
// scheme for an internal tool, not a public-facing product. See
// internal/auth's package doc.
const minPasswordLength = 8

type signupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Signup creates a new account for an allowlisted email and logs the user
// in immediately (sets the session cookie), matching Login's response
// shape.
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	email := normalizeEmail(req.Email)
	if email == "" || !strings.Contains(email, "@") {
		writeError(w, http.StatusBadRequest, "a valid email is required")
		return
	}
	if len(req.Password) < minPasswordLength {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("password must be at least %d characters", minPasswordLength))
		return
	}
	if !auth.Allowed(email, h.AllowedEmails) {
		writeError(w, http.StatusForbidden, "this email is not on the approved list for this team")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to process password")
		return
	}

	user := models.User{Email: email, PasswordHash: hash, CreatedAt: time.Now().UTC()}
	if _, err := h.DB.Users.InsertOne(ctx, user); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			writeError(w, http.StatusConflict, "an account with this email already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create account")
		return
	}

	h.startSession(w, r.Context(), email)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login verifies email/password and, on success, starts a new session.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	email := normalizeEmail(req.Email)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var user models.User
	switch err := h.DB.Users.FindOne(ctx, bson.M{"email": email}).Decode(&user); {
	case err == mongo.ErrNoDocuments:
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, "failed to look up account")
		return
	}

	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	h.startSession(w, r.Context(), email)
}

// startSession creates a session for email, sets the cookie, and writes
// the shared success response used by both Signup and Login.
func (h *Handler) startSession(w http.ResponseWriter, ctx context.Context, email string) {
	sessionCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cookie, err := h.AuthStore.Create(sessionCtx, email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start session")
		return
	}
	auth.SetSessionCookie(w, cookie, h.CookieSecure)

	writeJSON(w, http.StatusOK, map[string]string{"email": email})
}

// Logout clears the caller's session, both server-side and the cookie.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(auth.CookieName); err == nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		_ = h.AuthStore.Delete(ctx, cookie.Value)
	}
	auth.ClearSessionCookie(w, h.CookieSecure)
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// Me returns the authenticated caller's email. Must sit behind
// auth.RequireAuth.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"email": auth.UserEmail(r)})
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
