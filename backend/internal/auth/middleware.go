package auth

import (
	"context"
	"net/http"
	"time"
)

// RequireAuth wraps handlers that need a logged-in user. It reads the
// session cookie, validates it against store, and — on success — stores
// the user's email in the request context for handlers to read via
// UserEmail(r). A request with no cookie, or an invalid/expired one, gets
// a 401 before the wrapped handler ever runs.
func RequireAuth(store *Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(CookieName)
			if err != nil {
				writeUnauthorized(w, "not authenticated")
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			email, ok := store.Validate(ctx, cookie.Value)
			if !ok {
				writeUnauthorized(w, "session expired or invalid; please log in again")
				return
			}

			next.ServeHTTP(w, r.WithContext(WithUserEmail(r.Context(), email)))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"` + message + `"}`))
}
