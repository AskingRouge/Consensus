package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Comment is one reviewer's remark in an item's discussion thread.
type Comment struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ItemID      primitive.ObjectID `bson:"item_id" json:"item_id"`
	SessionID   primitive.ObjectID `bson:"session_id" json:"session_id"`
	AuthorEmail string             `bson:"author_email" json:"author_email"`
	Text        string             `bson:"text" json:"text"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

// CommentInput is the shape accepted from the client when posting a comment.
type CommentInput struct {
	Text string `json:"text"`
}
