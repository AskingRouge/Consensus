package ws

import (
	"encoding/json"
	"testing"
	"time"
)

// recvPresence reads one message from a client's outbound channel and
// decodes it as a presence snapshot, failing the test if nothing arrives
// in time or it isn't a presence message.
func recvPresence(t *testing.T, c *Client) []PresenceEntry {
	t.Helper()
	select {
	case data := <-c.send:
		var env envelope
		if err := json.Unmarshal(data, &env); err != nil {
			t.Fatalf("bad envelope: %v", err)
		}
		if env.Type != TypePresence {
			t.Fatalf("type = %q, want %q", env.Type, TypePresence)
		}
		var entries []PresenceEntry
		if err := json.Unmarshal(env.Payload, &entries); err != nil {
			t.Fatalf("bad payload: %v", err)
		}
		return entries
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for presence message")
		return nil
	}
}

func TestSessionHub_AddClientBroadcastsPresence(t *testing.T) {
	sh := newSessionHub("s1", func() {})
	c := newClient(sh, nil, "alice@unc.edu")

	sh.addClient(c)
	sh.broadcastPresence()

	entries := recvPresence(t, c)
	if len(entries) != 1 || entries[0].ReviewerEmail != "alice@unc.edu" {
		t.Errorf("unexpected presence: %+v", entries)
	}
}

func TestSessionHub_ViewItemReflectedInPresence(t *testing.T) {
	sh := newSessionHub("s1", func() {})
	c := newClient(sh, nil, "alice@unc.edu")
	sh.addClient(c)
	sh.broadcastPresence()
	recvPresence(t, c) // drain the initial snapshot

	c.setViewingItem("item-42")
	sh.broadcastPresence()

	entries := recvPresence(t, c)
	if len(entries) != 1 || entries[0].ItemID != "item-42" {
		t.Errorf("unexpected presence: %+v", entries)
	}
}

func TestSessionHub_DisconnectLastClientCallsOnEmpty(t *testing.T) {
	emptied := make(chan struct{}, 1)
	sh := newSessionHub("s1", func() { emptied <- struct{}{} })
	c := newClient(sh, nil, "alice@unc.edu")
	sh.addClient(c)

	sh.disconnect(c)

	select {
	case <-emptied:
	case <-time.After(time.Second):
		t.Fatal("onEmpty was not called after the last client disconnected")
	}

	if _, stillOpen := <-c.send; stillOpen {
		t.Error("client's send channel should be closed after disconnect")
	}
}

func TestSessionHub_DisconnectOneOfTwoBroadcastsUpdatedPresence(t *testing.T) {
	sh := newSessionHub("s1", func() { t.Fatal("onEmpty should not fire while a client remains") })
	a := newClient(sh, nil, "alice@unc.edu")
	b := newClient(sh, nil, "bob@unc.edu")
	sh.addClient(a)
	sh.addClient(b)

	sh.disconnect(a)

	entries := recvPresence(t, b)
	if len(entries) != 1 || entries[0].ReviewerEmail != "bob@unc.edu" {
		t.Errorf("expected only bob left in presence, got %+v", entries)
	}
}

func TestSessionHub_DisconnectIsIdempotent(t *testing.T) {
	calls := 0
	sh := newSessionHub("s1", func() { calls++ })
	c := newClient(sh, nil, "alice@unc.edu")
	sh.addClient(c)

	sh.disconnect(c)
	sh.disconnect(c) // must not panic (double-close) or call onEmpty twice

	if calls != 1 {
		t.Errorf("onEmpty called %d times, want 1", calls)
	}
}

func TestClient_TrySendDisconnectsWhenBufferFull(t *testing.T) {
	sh := newSessionHub("s1", func() {})
	c := newClient(sh, nil, "alice@unc.edu")
	sh.addClient(c)

	// Fill the buffer past capacity so the next send has to drop the
	// client instead of blocking the caller.
	for i := 0; i < sendBufferSize+1; i++ {
		c.trySend([]byte("x"))
	}

	select {
	case _, open := <-c.send:
		if open {
			// Draining is fine; what matters is the hub no longer
			// considers this client connected.
		}
	default:
	}

	if len(sh.snapshotClients()) != 0 {
		t.Error("a client whose buffer overflowed should have been disconnected")
	}
}

func TestHub_BroadcastToUnknownSessionIsNoop(t *testing.T) {
	h := NewHub()
	// Must not panic even though no client has ever connected.
	h.Broadcast("nonexistent", TypeScoreUpdate, map[string]string{"ok": "true"})
}

func TestHub_BroadcastReachesConnectedClient(t *testing.T) {
	h := NewHub()
	sh := h.getOrCreate("s1")
	c := newClient(sh, nil, "alice@unc.edu")
	sh.addClient(c)

	h.Broadcast("s1", TypeScoreUpdate, map[string]string{"item_id": "i1"})

	select {
	case data := <-c.send:
		var env envelope
		if err := json.Unmarshal(data, &env); err != nil {
			t.Fatalf("bad envelope: %v", err)
		}
		if env.Type != TypeScoreUpdate {
			t.Errorf("type = %q, want %q", env.Type, TypeScoreUpdate)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for broadcast")
	}
}

func TestHub_GetOrCreateReusesExistingHub(t *testing.T) {
	h := NewHub()
	a := h.getOrCreate("s1")
	b := h.getOrCreate("s1")
	if a != b {
		t.Error("getOrCreate should return the same hub for the same session id")
	}
}
