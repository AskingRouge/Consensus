package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"consensus/internal/auth"
	"consensus/internal/models"
)

// RequireSessionMember gates every /api/sessions/{sessionID}/* route
// behind session membership: the session must exist, and the
// authenticated caller must be in its Members list. This is the
// enforcement point for private/scoped sessions — being on the team's
// global allowlist gets a request past auth.RequireAuth, but only being a
// member of *this* session gets it past here. A non-member gets a 404
// rather than a 403, so a session's existence isn't leaked to people who
// aren't on it.
func (h *Handler) RequireSessionMember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID, ok := parseObjectID(w, chi.URLParam(r, "sessionID"))
		if !ok {
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var session models.Session
		switch err := h.DB.Sessions.FindOne(ctx, bson.M{"_id": sessionID}).Decode(&session); {
		case err == mongo.ErrNoDocuments:
			writeError(w, http.StatusNotFound, "session not found")
			return
		case err != nil:
			writeError(w, http.StatusInternalServerError, "failed to look up session")
			return
		}

		if !isMember(session, auth.UserEmail(r)) {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isMember(session models.Session, email string) bool {
	for _, m := range session.Members {
		if m == email {
			return true
		}
	}
	return false
}
