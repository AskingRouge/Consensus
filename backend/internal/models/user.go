package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is a team member with a login account. Only emails on the
// allowlist may sign up in the first place (see internal/auth.Allowed),
// so this collection is effectively the roster of active reviewers.
//
// PasswordHash is tagged json:"-" so it never round-trips into an API
// response, even by accident.
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"password_hash" json:"-"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}
