package auth

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// sessionTTL is how long a login lasts before it must be renewed. MongoDB
// enforces this server-side via a TTL index (see db.ensureIndexes) so an
// expired session disappears from the database on its own.
const sessionTTL = 7 * 24 * time.Hour

// sessionDoc is the auth_sessions document shape: an opaque token (as the
// document ID) mapped to the user it authenticates.
type sessionDoc struct {
	Token     string    `bson:"_id"`
	UserEmail string    `bson:"user_email"`
	ExpiresAt time.Time `bson:"expires_at"`
}

// Store manages signed, database-backed login sessions. It's the one
// place that knows how a cookie value maps to a logged-in user.
type Store struct {
	collection *mongo.Collection
	secret     string
}

// NewStore builds a Store over the given collection, signing/verifying
// cookie values with secret (config.Config.SessionSecret).
func NewStore(collection *mongo.Collection, secret string) *Store {
	return &Store{collection: collection, secret: secret}
}

// Create starts a new session for userEmail and returns the signed cookie
// value to hand back to the client via SetSessionCookie.
func (s *Store) Create(ctx context.Context, userEmail string) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}

	doc := sessionDoc{
		Token:     token,
		UserEmail: userEmail,
		ExpiresAt: time.Now().Add(sessionTTL),
	}
	if _, err := s.collection.InsertOne(ctx, doc); err != nil {
		return "", err
	}

	return sign(token, s.secret), nil
}

// Validate checks a signed cookie value and, if it corresponds to a live
// session, returns the session's user email.
func (s *Store) Validate(ctx context.Context, signedCookie string) (string, bool) {
	token, ok := verify(signedCookie, s.secret)
	if !ok {
		return "", false
	}

	var doc sessionDoc
	if err := s.collection.FindOne(ctx, bson.M{"_id": token}).Decode(&doc); err != nil {
		return "", false
	}
	if time.Now().After(doc.ExpiresAt) {
		return "", false
	}

	return doc.UserEmail, true
}

// Delete removes a session (logout). A cookie value that doesn't verify
// or doesn't exist is treated as already logged out, not an error.
func (s *Store) Delete(ctx context.Context, signedCookie string) error {
	token, ok := verify(signedCookie, s.secret)
	if !ok {
		return nil
	}
	_, err := s.collection.DeleteOne(ctx, bson.M{"_id": token})
	return err
}
