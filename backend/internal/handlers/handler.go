// Package handlers implements the REST API: session and item CRUD, score
// and comment submission, and export. Handlers are thin — they validate
// input, talk to db.DB, and (once wired up) notify the WebSocket hub of
// anything other connected clients need to know about.
package handlers

import "consensus/internal/db"

// Handler bundles the dependencies every REST handler needs. One instance
// is constructed in main.go and its methods are registered as routes.
type Handler struct {
	DB *db.DB
}

// New constructs a Handler.
func New(database *db.DB) *Handler {
	return &Handler{DB: database}
}
