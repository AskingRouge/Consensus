package handlers

import (
	"context"
	"net/http"
)

type contextKey string

const userEmailContextKey contextKey = "userEmail"

// WithUserEmail returns a context carrying the authenticated user's email.
// The auth middleware (internal/auth, wired up in Phase 2) sets this after
// verifying the Google OAuth session and checking the allowlist.
func WithUserEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, userEmailContextKey, email)
}

// UserEmail returns the authenticated user's email for this request.
//
// Until Phase 2 wires up Google OAuth, no middleware sets the context
// value, so this falls back to trusting an X-Dev-User-Email header — that
// keeps every handler's "who is making this request" logic identical
// before and after auth lands, instead of being rewritten in Phase 2. Once
// auth/middleware.go is in the chain, requests without a valid session
// never reach here at all.
func UserEmail(r *http.Request) string {
	if email, ok := r.Context().Value(userEmailContextKey).(string); ok && email != "" {
		return email
	}
	return r.Header.Get("X-Dev-User-Email")
}
