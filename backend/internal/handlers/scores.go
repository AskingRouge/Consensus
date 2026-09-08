package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"consensus/internal/auth"
	"consensus/internal/models"
	"consensus/internal/scoring"
	"consensus/internal/ws"
)

// ScoreUpdate is the shape returned by SubmitScore and broadcast to every
// connected client on ws.TypeScoreUpdate. Using one type for both means
// the REST response is fully sufficient on its own — a reviewer whose
// WebSocket happened to be disconnected at the moment they submitted
// still ends up with a correct, up-to-date view of the item's flag status
// from the HTTP response alone, with the WebSocket purely as a live
// convenience for everyone else.
type ScoreUpdate struct {
	Score                 models.Score `json:"score"`
	Flagged               bool         `json:"flagged"`
	DisagreeingDimensions []string     `json:"disagreeing_dimensions,omitempty"`
}

type submitScoreRequest struct {
	RubricScores map[string]float64 `json:"rubric_scores"`
}

// SubmitScore creates or updates the caller's score for one item (one
// score per reviewer per item — resubmitting overwrites, it doesn't add a
// second score), recomputes whether the item now needs discussion across
// every reviewer's score, persists that onto the item, and broadcasts the
// result to every client connected to the session.
func (h *Handler) SubmitScore(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := parseObjectID(w, chi.URLParam(r, "sessionID"))
	if !ok {
		return
	}
	itemID, ok := parseObjectID(w, chi.URLParam(r, "itemID"))
	if !ok {
		return
	}

	reviewer := auth.UserEmail(r)
	if reviewer == "" {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req submitScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var session models.Session
	switch err := h.DB.Sessions.FindOne(ctx, bson.M{"_id": sessionID}).Decode(&session); {
	case err == mongo.ErrNoDocuments:
		writeError(w, http.StatusNotFound, "session not found")
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, "failed to load session")
		return
	}

	if err := h.DB.Items.FindOne(ctx, bson.M{"_id": itemID, "session_id": sessionID}).Err(); err == mongo.ErrNoDocuments {
		writeError(w, http.StatusNotFound, "item not found in this session")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to look up item")
		return
	}

	validated, err := scoring.ValidateRubricScores(req.RubricScores, session.Rubric)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := bson.M{"item_id": itemID, "reviewer_email": reviewer}
	updateDoc := bson.M{"$set": bson.M{
		"item_id":        itemID,
		"session_id":     sessionID,
		"reviewer_email": reviewer,
		"rubric_scores":  validated,
		"submitted_at":   time.Now().UTC(),
	}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var score models.Score
	if err := h.DB.Scores.FindOneAndUpdate(ctx, filter, updateDoc, opts).Decode(&score); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save score")
		return
	}

	result, err := h.recomputeItemFlag(ctx, itemID, session.Rubric, session.DisagreementThreshold)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update disagreement flag")
		return
	}

	payload := ScoreUpdate{Score: score, Flagged: result.Flagged, DisagreeingDimensions: result.DisagreeingDimensions}
	h.Hub.Broadcast(sessionID.Hex(), ws.TypeScoreUpdate, payload)

	writeJSON(w, http.StatusOK, payload)
}

// recomputeItemFlag re-reads every score on itemID, runs disagreement
// detection, and persists the result onto the item document.
func (h *Handler) recomputeItemFlag(ctx context.Context, itemID primitive.ObjectID, rubric []models.RubricDimension, threshold float64) (scoring.Result, error) {
	cur, err := h.DB.Scores.Find(ctx, bson.M{"item_id": itemID})
	if err != nil {
		return scoring.Result{}, err
	}
	var allScores []models.Score
	if err := cur.All(ctx, &allScores); err != nil {
		return scoring.Result{}, err
	}

	result := scoring.Detect(allScores, rubric, threshold)

	_, err = h.DB.Items.UpdateOne(ctx, bson.M{"_id": itemID}, bson.M{"$set": bson.M{
		"flagged":                result.Flagged,
		"disagreeing_dimensions": result.DisagreeingDimensions,
	}})
	return result, err
}

// ListScores returns every score submitted anywhere in the session — the
// frontend groups these by item itself. Also doubles as the data source
// for export (Phase 7).
func (h *Handler) ListScores(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := parseObjectID(w, chi.URLParam(r, "sessionID"))
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	cur, err := h.DB.Scores.Find(ctx, bson.M{"session_id": sessionID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list scores")
		return
	}
	defer cur.Close(ctx)

	scores := []models.Score{}
	if err := cur.All(ctx, &scores); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read scores")
		return
	}

	writeJSON(w, http.StatusOK, scores)
}
