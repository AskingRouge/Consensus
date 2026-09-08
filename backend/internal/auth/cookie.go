package auth

import "net/http"

// CookieName is the session cookie sent to and read back from the browser.
const CookieName = "consensus_session"

// SetSessionCookie attaches a signed session cookie to the response.
// secure should be true whenever the frontend is served over HTTPS
// (anything other than local http://localhost dev) — browsers refuse to
// send a Secure cookie over plain HTTP, so getting this wrong in
// production silently breaks login rather than being insecure.
func SetSessionCookie(w http.ResponseWriter, value string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
}

// ClearSessionCookie removes the session cookie (logout).
func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
