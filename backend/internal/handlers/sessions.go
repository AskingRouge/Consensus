package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"consensus/internal/auth"
	"consensus/internal/models"
)

type createSessionRequest struct {
	Name                  string                   `json:"name"`
	Rubric                []models.RubricDimension `json:"rubric"`
	DisagreementThreshold *float64                 `json:"disagreement_threshold"`
}

// CreateSession creates a new review session. If no rubric is supplied,
// the default three-dimension rubric is used. The creator's identity comes
// from the authenticated request context — this route sits behind
// auth.RequireAuth, so UserEmail(r) is always populated.
func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	rubric := req.Rubric
	if len(rubric) == 0 {
		rubric = models.DefaultRubric()
	}
	for _, dim := range rubric {
		if dim.Name == "" {
			writeError(w, http.StatusBadRequest, "every rubric dimension needs a name")
			return
		}
		if dim.Type != models.RubricTypeScale && dim.Type != models.RubricTypePassFail {
			writeError(w, http.StatusBadRequest, "rubric dimension type must be \"scale\" or \"pass_fail\"")
			return
		}
		if dim.Type == models.RubricTypeScale && dim.ScaleMax <= 0 {
			writeError(w, http.StatusBadRequest, "scale rubric dimensions need a positive scale_max")
			return
		}
	}

	threshold := models.DefaultDisagreementThreshold
	if req.DisagreementThreshold != nil {
		threshold = *req.DisagreementThreshold
	}

	createdBy := auth.UserEmail(r)
	if createdBy == "" {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	session := models.Session{
		Name:                  req.Name,
		CreatedAt:             time.Now().UTC(),
		CreatedBy:             createdBy,
		Rubric:                rubric,
		DisagreementThreshold: threshold,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	res, err := h.DB.Sessions.InsertOne(ctx, session)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	session.ID = res.InsertedID.(primitive.ObjectID)

	writeJSON(w, http.StatusCreated, session)
}

// ListSessions returns every session, newest first.
func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, err := h.DB.Sessions.Find(ctx, bson.M{}, opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list sessions")
		return
	}
	defer cur.Close(ctx)

	sessions := []models.Session{}
	if err := cur.All(ctx, &sessions); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read sessions")
		return
	}

	writeJSON(w, http.StatusOK, sessions)
}

// GetSession returns a single session by ID.
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	id, ok := parseObjectID(w, chi.URLParam(r, "sessionID"))
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var session models.Session
	err := h.DB.Sessions.FindOne(ctx, bson.M{"_id": id}).Decode(&session)
	if err == mongo.ErrNoDocuments {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load session")
		return
	}

	writeJSON(w, http.StatusOK, session)
}
