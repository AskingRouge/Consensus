package handlers

import (
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// parseObjectID parses a Mongo ObjectID from a URL path parameter, writing
// a 400 response and returning ok=false if it isn't a valid ObjectID hex
// string.
func parseObjectID(w http.ResponseWriter, raw string) (primitive.ObjectID, bool) {
	id, err := primitive.ObjectIDFromHex(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return primitive.NilObjectID, false
	}
	return id, true
}
