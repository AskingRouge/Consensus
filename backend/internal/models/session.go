// Package models defines the MongoDB-backed domain types shared across the
// backend: sessions, items, scores, and comments.
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RubricDimensionType identifies how a rubric dimension is scored.
type RubricDimensionType string

const (
	RubricTypeScale    RubricDimensionType = "scale"
	RubricTypePassFail RubricDimensionType = "pass_fail"
)

// RubricDimension is one axis of evaluation within a session's rubric, e.g.
// "Retrieval Quality" scored 1-5, or "Safety/Appropriateness" as pass/fail.
type RubricDimension struct {
	Name     string              `bson:"name" json:"name"`
	Type     RubricDimensionType `bson:"type" json:"type"`
	ScaleMax int                 `bson:"scale_max,omitempty" json:"scale_max,omitempty"`
}

// Session represents one batch of LLM outputs to review.
//
// Members is the access-control list: only these emails (always including
// CreatedBy, added at creation) can see or act on this session — it's
// excluded from ListSessions results and every other session-scoped
// endpoint for anyone not on it. Being on the global ALLOWED_EMAILS
// allowlist gets you into the team; being a session Member gets you into
// that specific session.
type Session struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name                  string             `bson:"name" json:"name"`
	CreatedAt             time.Time          `bson:"created_at" json:"created_at"`
	CreatedBy             string             `bson:"created_by" json:"created_by"`
	Rubric                []RubricDimension  `bson:"rubric" json:"rubric"`
	DisagreementThreshold float64            `bson:"disagreement_threshold" json:"disagreement_threshold"`
	Members               []string           `bson:"members" json:"members"`
}

// DefaultRubric returns the rubric shipped by default, matching the UNC
// research harness output format: Retrieval Quality (1-5), Course-Scope
// Adherence (1-5), and Safety/Appropriateness (pass/fail).
func DefaultRubric() []RubricDimension {
	return []RubricDimension{
		{Name: "Retrieval Quality", Type: RubricTypeScale, ScaleMax: 5},
		{Name: "Course-Scope Adherence", Type: RubricTypeScale, ScaleMax: 5},
		{Name: "Safety/Appropriateness", Type: RubricTypePassFail},
	}
}

// DefaultDisagreementThreshold is the score-delta (on a scale dimension)
// above which an item is flagged as needing discussion.
const DefaultDisagreementThreshold = 2.0
