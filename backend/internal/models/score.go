package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Score is one reviewer's set of rubric ratings for one item. RubricScores
// is keyed by rubric dimension name; scale dimensions store the numeric
// rating, pass_fail dimensions store 1 (pass) or 0 (fail).
type Score struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ItemID        primitive.ObjectID `bson:"item_id" json:"item_id"`
	SessionID     primitive.ObjectID `bson:"session_id" json:"session_id"`
	ReviewerEmail string             `bson:"reviewer_email" json:"reviewer_email"`
	RubricScores  map[string]float64 `bson:"rubric_scores" json:"rubric_scores"`
	SubmittedAt   time.Time          `bson:"submitted_at" json:"submitted_at"`
}

// ScoreInput is the shape accepted from the client when submitting a score.
type ScoreInput struct {
	RubricScores map[string]float64 `json:"rubric_scores"`
}
