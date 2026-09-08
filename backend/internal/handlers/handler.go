// Package handlers implements the REST API: auth, session and item CRUD,
// score and comment submission, and export. Handlers are thin — they
// validate input, talk to db.DB and internal/auth, and (starting Phase 5)
// call Hub.Broadcast so other connected clients see the change live.
package handlers

import (
	"consensus/internal/auth"
	"consensus/internal/db"
	"consensus/internal/ws"
)

// Handler bundles the dependencies every REST handler needs. One instance
// is constructed in main.go and its methods are registered as routes.
type Handler struct {
	DB        *db.DB
	AuthStore *auth.Store
	// Hub lets handlers push a live update after a successful write —
	// e.g. CreateScore (Phase 5) broadcasting ws.TypeScoreUpdate. Unused
	// until then, but wired up now so that phase is a small addition
	// rather than a restructuring.
	Hub *ws.Hub

	// AllowedEmails is the signup allowlist (config.Config.AllowedEmails).
	AllowedEmails []string
	// CookieSecure controls the Secure flag on the session cookie — true
	// once the frontend is served over HTTPS.
	CookieSecure bool
}

// New constructs a Handler.
func New(database *db.DB, authStore *auth.Store, hub *ws.Hub, allowedEmails []string, cookieSecure bool) *Handler {
	return &Handler{
		DB:            database,
		AuthStore:     authStore,
		Hub:           hub,
		AllowedEmails: allowedEmails,
		CookieSecure:  cookieSecure,
	}
}
