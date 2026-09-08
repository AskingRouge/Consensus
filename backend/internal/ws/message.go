// Package ws is the WebSocket hub: one lightweight "room" per review
// session, tracking which reviewers are connected, broadcasting presence
// as they move between items, and (starting Phase 5) relaying score and
// comment updates the REST handlers write to Mongo out to every other
// connected client in real time.
//
// Every outbound message is a full snapshot, not a delta (e.g. "here is
// everyone currently viewing something" rather than "reviewer X started
// viewing item Y"). That's deliberate: a client that just reconnected
// after a dropped connection can render correctly from the next message
// it receives without needing to replay history it may have missed.
package ws

import "encoding/json"

// envelope is the wire format for every message the hub sends: a type tag
// plus a type-specific payload.
type envelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// encode wraps payload in an envelope tagged with msgType.
func encode(msgType string, payload interface{}) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(envelope{Type: msgType, Payload: raw})
}

// Outbound message types.
const (
	// TypePresence carries a full []PresenceEntry snapshot of who is
	// currently connected to the session and which item (if any) each
	// reviewer is looking at.
	TypePresence = "presence"

	// TypeScoreUpdate carries a models.Score — broadcast whenever any
	// reviewer submits or changes a score (wired up in Phase 5).
	TypeScoreUpdate = "score_update"

	// TypeComment carries a models.Comment — broadcast whenever a new
	// comment is posted (wired up in Phase 6).
	TypeComment = "comment"

	// TypeError carries {"message": string} for a problem with an
	// inbound message (e.g. malformed JSON) that doesn't warrant closing
	// the connection.
	TypeError = "error"
)

// PresenceEntry describes one connected reviewer's current focus.
type PresenceEntry struct {
	ReviewerEmail string `json:"reviewer_email"`
	// ItemID is the item this reviewer is currently viewing, or "" if
	// they're on the session's item list rather than a specific item.
	ItemID string `json:"item_id,omitempty"`
}

// inboundMessage is the shape a client is expected to send. Only
// "view_item" exists today; unrecognized types are ignored rather than
// closing the connection, so a slightly newer/older client doesn't get
// disconnected over a protocol mismatch.
type inboundMessage struct {
	Type   string `json:"type"`
	ItemID string `json:"item_id"`
}

// TypeViewItem: a reviewer navigated to (or away from, with ItemID="")
// a specific item.
const TypeViewItem = "view_item"
