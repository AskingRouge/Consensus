package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"consensus/internal/export"
	"consensus/internal/models"
)

// ExportSession renders every item, score, and comment in a session as
// either JSON (?format=json, the default) or CSV (?format=csv) so the
// research team can pull results back into their evaluation pipeline.
func (h *Handler) ExportSession(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := parseObjectID(w, chi.URLParam(r, "sessionID"))
	if !ok {
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}
	if format != "json" && format != "csv" {
		writeError(w, http.StatusBadRequest, `format must be "json" or "csv"`)
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

	items, err := findAll[models.Item](ctx, h.DB.Items, bson.M{"session_id": sessionID}, options.Find().SetSort(bson.D{{Key: "order_index", Value: 1}}))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load items")
		return
	}

	scores, err := findAll[models.Score](ctx, h.DB.Scores, bson.M{"session_id": sessionID})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load scores")
		return
	}

	comments, err := findAll[models.Comment](ctx, h.DB.Comments, bson.M{"session_id": sessionID}, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load comments")
		return
	}

	data := export.Data{Session: session, Items: items, Scores: scores, Comments: comments}
	filenameBase := fmt.Sprintf("consensus-export-%s", sessionID.Hex())

	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.csv"`, filenameBase))
		if err := export.WriteCSV(w, data); err != nil {
			log.Printf("export: failed to write CSV for session %s: %v", sessionID.Hex(), err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.json"`, filenameBase))
	if err := export.WriteJSON(w, data); err != nil {
		log.Printf("export: failed to write JSON for session %s: %v", sessionID.Hex(), err)
	}
}

// findAll runs a Mongo find and decodes every result into a slice,
// returning [] rather than nil when there are no matches.
func findAll[T any](ctx context.Context, collection *mongo.Collection, filter bson.M, opts ...*options.FindOptions) ([]T, error) {
	cur, err := collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := []T{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
