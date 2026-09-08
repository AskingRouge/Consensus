package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// writeWait is how long a single WriteMessage call may take.
	writeWait = 10 * time.Second
	// pongWait is how long we'll wait for a pong (or any message) before
	// considering the connection dead.
	pongWait = 60 * time.Second
	// pingPeriod must be less than pongWait; the standard 90% margin.
	pingPeriod = (pongWait * 9) / 10
	// maxMessageSize bounds inbound messages — the only client messages
	// today are tiny "view_item" notices, so this is generous, not tight.
	maxMessageSize = 4096
	// sendBufferSize is how many outbound messages can queue for a client
	// before it's considered too slow and disconnected. Presence/score/
	// comment events are small and infrequent, so this rarely matters.
	sendBufferSize = 32
)

// Client is one reviewer's live WebSocket connection. gorilla/websocket
// connections may not be written to from multiple goroutines at once, so
// all writes go through the single writePump goroutine via send; every
// other goroutine that wants to push data to this client uses trySend.
type Client struct {
	hub  *sessionHub
	conn *websocket.Conn
	send chan []byte

	email string

	mu            sync.Mutex
	currentItemID string

	closeOnce sync.Once
}

func newClient(hub *sessionHub, conn *websocket.Conn, email string) *Client {
	return &Client{
		hub:   hub,
		conn:  conn,
		send:  make(chan []byte, sendBufferSize),
		email: email,
	}
}

func (c *Client) viewingItem() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.currentItemID
}

func (c *Client) setViewingItem(itemID string) {
	c.mu.Lock()
	c.currentItemID = itemID
	c.mu.Unlock()
}

// trySend enqueues data for delivery without blocking. A client whose
// buffer is already full is disconnected rather than allowed to
// back-pressure every broadcast to everyone else in the session.
func (c *Client) trySend(data []byte) {
	select {
	case c.send <- data:
	default:
		c.hub.disconnect(c)
	}
}

// close is safe to call more than once or concurrently.
func (c *Client) close() {
	c.closeOnce.Do(func() {
		close(c.send)
	})
}

// readPump reads inbound messages until the connection closes, then
// unregisters the client from its hub. Run this in its own goroutine.
func (c *Client) readPump() {
	defer c.hub.disconnect(c)

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		var msg inboundMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue // ignore malformed messages rather than dropping the connection
		}

		switch msg.Type {
		case TypeViewItem:
			c.setViewingItem(msg.ItemID)
			c.hub.broadcastPresence()
		}
	}
}

// writePump drains send and writes to the connection, plus periodic
// pings to detect a dead connection before pongWait would. Run this in
// its own goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case data, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
