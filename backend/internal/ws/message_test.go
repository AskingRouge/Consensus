package ws

import (
	"encoding/json"
	"testing"
)

func TestEncode(t *testing.T) {
	data, err := encode(TypePresence, []PresenceEntry{{ReviewerEmail: "a@unc.edu", ItemID: "item1"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("failed to decode envelope: %v", err)
	}
	if env.Type != TypePresence {
		t.Errorf("type = %q, want %q", env.Type, TypePresence)
	}

	var entries []PresenceEntry
	if err := json.Unmarshal(env.Payload, &entries); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}
	if len(entries) != 1 || entries[0].ReviewerEmail != "a@unc.edu" || entries[0].ItemID != "item1" {
		t.Errorf("unexpected entries: %+v", entries)
	}
}
