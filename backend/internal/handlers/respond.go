package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// writeJSON encodes v as JSON with the given status code. It logs (rather
// than panics) if encoding fails after headers are already sent.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("handlers: failed to write JSON response: %v", err)
	}
}

// apiError is the JSON shape returned for any non-2xx response.
type apiError struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, apiError{Error: message})
}
