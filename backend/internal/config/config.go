// Package config loads process configuration from environment variables
// (optionally via a local .env file). Every value the app needs to run —
// database connection, OAuth credentials, allowlist, secrets — lives here
// so nothing is hardcoded, per the project's non-functional requirements.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds every environment-derived setting the server needs.
type Config struct {
	Port        string
	MongoURI    string
	MongoDB     string
	FrontendURL string

	// Google OAuth — not wired up yet (username/password auth is used for
	// now; see internal/auth). Left blank until that lands.
	GoogleOAuthClientID     string
	GoogleOAuthClientSecret string

	AllowedEmails []string
	SessionSecret string
	// SessionSecretGenerated is true when no SESSION_SECRET was configured
	// and Load() generated a random one for this process. That's fine for
	// a single local dev instance, but main.go logs a loud warning because
	// it means every login session is invalidated on restart, and it would
	// break auth entirely across multiple instances (e.g. once deployed).
	SessionSecretGenerated bool
}

// Load reads a .env file if present (local dev convenience; production
// deployments set real environment variables instead) and builds a Config,
// applying sane defaults for anything optional.
func Load() (Config, error) {
	// Ignore a missing .env — that's the normal case in production.
	_ = godotenv.Load()

	cfg := Config{
		Port:                    getEnv("PORT", "8080"),
		MongoURI:                getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDB:                 getEnv("MONGODB_DB", "consensus"),
		FrontendURL:             getEnv("FRONTEND_URL", "http://localhost:5173"),
		GoogleOAuthClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		GoogleOAuthClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		AllowedEmails:           splitAndTrim(os.Getenv("ALLOWED_EMAILS")),
		SessionSecret:           os.Getenv("SESSION_SECRET"),
	}

	if cfg.MongoURI == "" {
		return cfg, fmt.Errorf("MONGODB_URI must not be empty")
	}

	if cfg.SessionSecret == "" {
		secret, err := randomSecret()
		if err != nil {
			return cfg, fmt.Errorf("failed to generate a fallback SESSION_SECRET: %w", err)
		}
		cfg.SessionSecret = secret
		cfg.SessionSecretGenerated = true
	}

	return cfg, nil
}

func randomSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func splitAndTrim(csv string) []string {
	if csv == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
