// Package auth implements username/password login: password hashing, an
// email allowlist, and database-backed sessions delivered via a signed
// cookie. This is a deliberately temporary auth strategy — it's isolated
// behind this package specifically so it can be swapped for Google OAuth
// later without the rest of the app (handlers, frontend) changing shape:
// callers only ever see auth.UserEmail(r) and auth.RequireAuth(store).
package auth

import "golang.org/x/crypto/bcrypt"

// bcryptCost trades off hashing time against resistance to brute force.
// 12 is a reasonable default as of 2024-2025 hardware.
const bcryptCost = 12

// HashPassword produces a bcrypt hash suitable for storing in
// models.User.PasswordHash. The salt is embedded in the returned hash, so
// nothing else needs to be stored alongside it.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword reports whether password matches a hash previously
// produced by HashPassword.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
