package ws

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"consensus/internal/auth"
	"consensus/internal/db"
	"consensus/internal/models"
)

// Handler upgrades HTTP requests to WebSocket connections and joins them
// to the right session's hub.
type Handler struct {
	Hub *Hub
	DB  *db.DB

	// AllowedOrigin is checked against the WebSocket handshake's Origin
	// header to prevent another site from opening a WebSocket using a
	// visitor's browser session (cross-site WebSocket hijacking). It's
	// config.Config.FrontendURL — the same origin the REST API's CORS
	// policy already trusts.
	AllowedOrigin string
}

// NewHandler constructs a Handler.
func NewHandler(hub *Hub, database *db.DB, allowedOrigin string) *Handler {
	return &Handler{Hub: hub, DB: database, AllowedOrigin: allowedOrigin}
}

func (h *Handler) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	// No Origin header at all means a non-browser client (curl, a test) —
	// allow it; browsers always send Origin for cross-origin WebSocket
	// handshakes, so this doesn't weaken the browser-facing protection.
	return origin == "" || origin == h.AllowedOrigin
}

// ServeSession upgrades the connection and joins the caller to
// sessionID's live hub. It must sit behind auth.RequireAuth — the
// reviewer's identity comes from the same session cookie the REST API
// uses, so there's no separate WebSocket-specific login step.
func (h *Handler) ServeSession(w http.ResponseWriter, r *http.Request) {
	email := auth.UserEmail(r)
	if email == "" {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	sessionID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "sessionID"))
	if err != nil {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var session models.Session
	switch err := h.DB.Sessions.FindOne(ctx, bson.M{"_id": sessionID}).Decode(&session); {
	case err == mongo.ErrNoDocuments:
		http.Error(w, "session not found", http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, "failed to look up session", http.StatusInternalServerError)
		return
	}
	if !isMember(session, email) {
		// 404, not 403 — matches the REST API's RequireSessionMember,
		// which doesn't reveal a session exists to non-members.
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     h.checkOrigin,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade already wrote an HTTP error response on failure.
		log.Printf("ws: upgrade failed for session %s: %v", sessionID.Hex(), err)
		return
	}

	sh := h.Hub.getOrCreate(sessionID.Hex())
	client := newClient(sh, conn, email)
	sh.addClient(client)
	sh.broadcastPresence()

	go client.writePump()
	go client.readPump()
}

func isMember(session models.Session, email string) bool {
	for _, m := range session.Members {
		if m == email {
			return true
		}
	}
	return false
}
