package ws

import (
	"log"
	"sync"
)

// sessionHub is the set of clients currently connected to one review
// session. It's created on the first connection and torn down once the
// last client leaves — an idle session costs nothing.
type sessionHub struct {
	id string

	mu      sync.Mutex
	clients map[*Client]struct{}

	// onEmpty is called (at most once) when the last client disconnects,
	// so the owning Hub can drop this sessionHub from its map.
	onEmpty func()
}

func newSessionHub(id string, onEmpty func()) *sessionHub {
	return &sessionHub{
		id:      id,
		clients: make(map[*Client]struct{}),
		onEmpty: onEmpty,
	}
}

func (sh *sessionHub) addClient(c *Client) {
	sh.mu.Lock()
	sh.clients[c] = struct{}{}
	sh.mu.Unlock()
}

// disconnect removes c from the hub, closes its send channel, and
// notifies remaining clients of the presence change — or, if c was the
// last client, tears the hub down via onEmpty. Safe to call more than
// once for the same client.
func (sh *sessionHub) disconnect(c *Client) {
	sh.mu.Lock()
	_, existed := sh.clients[c]
	delete(sh.clients, c)
	empty := len(sh.clients) == 0
	sh.mu.Unlock()

	if !existed {
		return
	}
	c.close()

	if empty {
		sh.onEmpty()
		return
	}
	sh.broadcastPresence()
}

func (sh *sessionHub) snapshotClients() []*Client {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	out := make([]*Client, 0, len(sh.clients))
	for c := range sh.clients {
		out = append(out, c)
	}
	return out
}

// broadcast sends msgType/payload to every currently-connected client.
func (sh *sessionHub) broadcast(msgType string, payload interface{}) {
	data, err := encode(msgType, payload)
	if err != nil {
		log.Printf("ws: failed to encode %s message: %v", msgType, err)
		return
	}
	for _, c := range sh.snapshotClients() {
		c.trySend(data)
	}
}

// broadcastPresence sends a full presence snapshot to everyone in the
// session — see the package doc for why this is a snapshot, not a delta.
func (sh *sessionHub) broadcastPresence() {
	clients := sh.snapshotClients()
	entries := make([]PresenceEntry, 0, len(clients))
	for _, c := range clients {
		entries = append(entries, PresenceEntry{
			ReviewerEmail: c.email,
			ItemID:        c.viewingItem(),
		})
	}
	sh.broadcast(TypePresence, entries)
}

// Hub owns one sessionHub per review session with connected clients.
// There's exactly one Hub per server process, constructed in main.go and
// shared between the WebSocket upgrade handler and (from Phase 5 on) the
// REST handlers that need to broadcast score/comment updates.
type Hub struct {
	mu       sync.Mutex
	sessions map[string]*sessionHub
}

// NewHub constructs an empty Hub.
func NewHub() *Hub {
	return &Hub{sessions: make(map[string]*sessionHub)}
}

// getOrCreate returns the sessionHub for sessionID, creating it if this is
// the first client to connect.
func (h *Hub) getOrCreate(sessionID string) *sessionHub {
	h.mu.Lock()
	defer h.mu.Unlock()

	if sh, ok := h.sessions[sessionID]; ok {
		return sh
	}

	var sh *sessionHub
	sh = newSessionHub(sessionID, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		// Only remove if this is still the hub instance we created — a
		// fresh connection may have already replaced it after this one
		// went idle, and we must not delete that newer hub.
		if h.sessions[sessionID] == sh {
			delete(h.sessions, sessionID)
		}
	})
	h.sessions[sessionID] = sh
	return sh
}

// Broadcast pushes msgType/payload to every client currently connected to
// sessionID. REST handlers call this after a successful write to Mongo
// (score submission, comments) so every other open tab sees the change
// live. If nobody is connected to that session right now, this is a
// no-op — Mongo remains the source of truth regardless; the WebSocket
// push is purely a live-update convenience on top of it.
func (h *Hub) Broadcast(sessionID string, msgType string, payload interface{}) {
	h.mu.Lock()
	sh, ok := h.sessions[sessionID]
	h.mu.Unlock()
	if !ok {
		return
	}
	sh.broadcast(msgType, payload)
}
