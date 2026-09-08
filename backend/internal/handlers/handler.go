// Package handlers implements the REST API: auth, session and item CRUD,
// score and comment submission, and export. Handlers are thin — they
// validate input, talk to db.DB and internal/auth, and (once wired up)
// notify the WebSocket hub of anything other connected clients need to
// know about.
package handlers

import (
	"consensus/internal/auth"
	"consensus/internal/db"
)

// Handler bundles the dependencies every REST handler needs. One instance
// is constructed in main.go and its methods are registered as routes.
type Handler struct {
	DB        *db.DB
	AuthStore *auth.Store

	// AllowedEmails is the signup allowlist (config.Config.AllowedEmails).
	AllowedEmails []string
	// CookieSecure controls the Secure flag on the session cookie — true
	// once the frontend is served over HTTPS.
	CookieSecure bool
}

// New constructs a Handler.
func New(database *db.DB, authStore *auth.Store, allowedEmails []string, cookieSecure bool) *Handler {
	return &Handler{
		DB:            database,
		AuthStore:     authStore,
		AllowedEmails: allowedEmails,
		CookieSecure:  cookieSecure,
	}
}
