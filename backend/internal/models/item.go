package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Item is a single LLM output to be reviewed: a question, the context that
// was retrieved for it, and the model's generated answer.
//
// Flagged and DisagreeingDimensions are derived data, recomputed by
// internal/scoring.Detect every time a score is submitted for this item
// (see handlers.SubmitScore) — they're stored rather than computed on
// every read so listing a session's items doesn't require an extra query
// per item just to know which ones need discussion.
type Item struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SessionID             primitive.ObjectID `bson:"session_id" json:"session_id"`
	Question              string             `bson:"question" json:"question"`
	RetrievedContext      string             `bson:"retrieved_context" json:"retrieved_context"`
	GeneratedAnswer       string             `bson:"generated_answer" json:"generated_answer"`
	OrderIndex            int                `bson:"order_index" json:"order_index"`
	Flagged               bool               `bson:"flagged" json:"flagged"`
	DisagreeingDimensions []string           `bson:"disagreeing_dimensions,omitempty" json:"disagreeing_dimensions,omitempty"`
}

// ItemInput is the shape accepted from an uploaded JSON/CSV batch, before a
// session ID and order index have been assigned.
type ItemInput struct {
	Question         string `json:"question"`
	RetrievedContext string `json:"retrieved_context"`
	GeneratedAnswer  string `json:"generated_answer"`
}
