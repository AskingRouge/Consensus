package auth

import "strings"

// Allowed reports whether email is permitted to sign up / log in. An empty
// allowlist denies everyone — this fails closed, not open, so a
// misconfigured ALLOWED_EMAILS locks everyone out rather than opening the
// tool to the public.
func Allowed(email string, allowlist []string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return false
	}
	for _, a := range allowlist {
		if a == email {
			return true
		}
	}
	return false
}
