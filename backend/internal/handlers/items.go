package handlers

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"consensus/internal/models"
	"consensus/internal/parse"
)

// maxUploadSize caps in-memory parsing of an uploaded item batch at 20MB.
const maxUploadSize = 20 << 20

// UploadItems accepts a multipart/form-data upload with a single "file"
// field containing a JSON or CSV batch of items (question,
// retrieved_context, generated_answer — the UNC research harness's output
// format) and inserts them into the session, ordered as given. The format
// is inferred from the file extension unless an explicit "format" query
// parameter ("json" or "csv") is supplied.
func (h *Handler) UploadItems(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := parseObjectID(w, chi.URLParam(r, "sessionID"))
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := h.DB.Sessions.FindOne(ctx, bson.M{"_id": sessionID}).Err(); err == mongo.ErrNoDocuments {
		writeError(w, http.StatusNotFound, "session not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to look up session")
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse upload (max 20MB)")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing \"file\" field in upload")
		return
	}
	defer file.Close()

	format := r.URL.Query().Get("format")
	if format == "" {
		switch strings.ToLower(filepath.Ext(header.Filename)) {
		case ".json":
			format = "json"
		case ".csv":
			format = "csv"
		default:
			writeError(w, http.StatusBadRequest, "cannot infer format from filename; pass ?format=json or ?format=csv")
			return
		}
	}

	inputs, err := parse.Items(file, format)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(inputs) == 0 {
		writeError(w, http.StatusBadRequest, "upload contained no items")
		return
	}

	// Items are appended after whatever is already in the session, so
	// re-uploading a second batch doesn't clobber the first.
	startIndex, err := h.nextOrderIndex(ctx, sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to determine item order")
		return
	}

	docs := make([]interface{}, len(inputs))
	items := make([]models.Item, len(inputs))
	for i, in := range inputs {
		item := models.Item{
			SessionID:        sessionID,
			Question:         in.Question,
			RetrievedContext: in.RetrievedContext,
			GeneratedAnswer:  in.GeneratedAnswer,
			OrderIndex:       startIndex + i,
		}
		items[i] = item
		docs[i] = item
	}

	res, err := h.DB.Items.InsertMany(ctx, docs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save items")
		return
	}
	for i, id := range res.InsertedIDs {
		items[i].ID = id.(primitive.ObjectID)
	}

	writeJSON(w, http.StatusCreated, items)
}

func (h *Handler) nextOrderIndex(ctx context.Context, sessionID primitive.ObjectID) (int, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "order_index", Value: -1}})
	var last models.Item
	err := h.DB.Items.FindOne(ctx, bson.M{"session_id": sessionID}, opts).Decode(&last)
	if err == mongo.ErrNoDocuments {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return last.OrderIndex + 1, nil
}

// ListItems returns every item in a session, in display order.
func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := parseObjectID(w, chi.URLParam(r, "sessionID"))
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "order_index", Value: 1}})
	cur, err := h.DB.Items.Find(ctx, bson.M{"session_id": sessionID}, opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list items")
		return
	}
	defer cur.Close(ctx)

	items := []models.Item{}
	if err := cur.All(ctx, &items); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read items")
		return
	}

	writeJSON(w, http.StatusOK, items)
}
