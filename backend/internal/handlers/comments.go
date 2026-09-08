package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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
	"consensus/internal/ws"
)

// maxCommentLength keeps a single comment from growing unboundedly large;
// generous for genuine discussion, not a place to paste a whole document.
const maxCommentLength = 4000

type createCommentRequest struct {
	Text string `json:"text"`
}

// CreateComment posts a new comment to an item's discussion thread and
// broadcasts it to every client connected to the session.
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := parseObjectID(w, chi.URLParam(r, "sessionID"))
	if !ok {
		return
	}
	itemID, ok := parseObjectID(w, chi.URLParam(r, "itemID"))
	if !ok {
		return
	}

	author := auth.UserEmail(r)
	if author == "" {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req createCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	text := strings.TrimSpace(req.Text)
	if text == "" {
		writeError(w, http.StatusBadRequest, "comment text is required")
		return
	}
	if len(text) > maxCommentLength {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("comment must be %d characters or fewer", maxCommentLength))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.DB.Items.FindOne(ctx, bson.M{"_id": itemID, "session_id": sessionID}).Err(); err == mongo.ErrNoDocuments {
		writeError(w, http.StatusNotFound, "item not found in this session")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to look up item")
		return
	}

	comment := models.Comment{
		ItemID:      itemID,
		SessionID:   sessionID,
		AuthorEmail: author,
		Text:        text,
		CreatedAt:   time.Now().UTC(),
	}

	res, err := h.DB.Comments.InsertOne(ctx, comment)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save comment")
		return
	}
	comment.ID = res.InsertedID.(primitive.ObjectID)

	h.Hub.Broadcast(sessionID.Hex(), ws.TypeComment, comment)

	writeJSON(w, http.StatusCreated, comment)
}

// ListComments returns every comment in the session, oldest first — the
// frontend groups these by item.
func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := parseObjectID(w, chi.URLParam(r, "sessionID"))
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	cur, err := h.DB.Comments.Find(ctx, bson.M{"session_id": sessionID}, opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list comments")
		return
	}
	defer cur.Close(ctx)

	comments := []models.Comment{}
	if err := cur.All(ctx, &comments); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read comments")
		return
	}

	writeJSON(w, http.StatusOK, comments)
}
