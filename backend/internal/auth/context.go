package auth

import (
	"context"
	"net/http"
)

type contextKey string

const userEmailContextKey contextKey = "userEmail"

// WithUserEmail returns a context carrying the authenticated user's email.
// RequireAuth sets this after validating the session cookie; handlers
// never set it themselves.
func WithUserEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, userEmailContextKey, email)
}

// UserEmail returns the authenticated user's email for this request, or
// "" if the request never passed through RequireAuth (which should only
// happen for a handler that was mistakenly left off a protected route —
// callers should treat "" as "not authenticated" and fail closed).
func UserEmail(r *http.Request) string {
	email, _ := r.Context().Value(userEmailContextKey).(string)
	return email
}
