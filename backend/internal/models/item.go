package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Item is a single LLM output to be reviewed: a question, the context that
// was retrieved for it, and the model's generated answer.
type Item struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SessionID        primitive.ObjectID `bson:"session_id" json:"session_id"`
	Question         string             `bson:"question" json:"question"`
	RetrievedContext string             `bson:"retrieved_context" json:"retrieved_context"`
	GeneratedAnswer  string             `bson:"generated_answer" json:"generated_answer"`
	OrderIndex       int                `bson:"order_index" json:"order_index"`
}

// ItemInput is the shape accepted from an uploaded JSON/CSV batch, before a
// session ID and order index have been assigned.
type ItemInput struct {
	Question         string `json:"question"`
	RetrievedContext string `json:"retrieved_context"`
	GeneratedAnswer  string `json:"generated_answer"`
}
