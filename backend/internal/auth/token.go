package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

// newToken generates a random, URL-safe opaque session token with 256 bits
// of entropy — this is the value stored (as the document ID) in the
// auth_sessions collection.
func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// sign produces "<token>.<hmac>" using SESSION_SECRET, so a cookie value
// can be rejected as forged or tampered-with before it ever reaches the
// database — the DB lookup only happens for values that were genuinely
// issued by this server.
func sign(token, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(token))
	return token + "." + hex.EncodeToString(mac.Sum(nil))
}

// verify checks a cookie value's signature and, if valid, returns the
// bare token to look up in the session store.
func verify(signed, secret string) (string, bool) {
	parts := strings.SplitN(signed, ".", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "", false
	}
	token := parts[0]
	expected := sign(token, secret)
	if subtle.ConstantTimeCompare([]byte(signed), []byte(expected)) != 1 {
		return "", false
	}
	return token, true
}
