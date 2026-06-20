package internal

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ── Event types ──

type EventType string

const (
	EventProjectCreated EventType = "project_created"
	EventProjectUpdated EventType = "project_updated"
	EventProjectDeleted EventType = "project_deleted"
	EventIssueCreated   EventType = "issue_created"
	EventIssueUpdated   EventType = "issue_updated"
	EventIssueDeleted   EventType = "issue_deleted"
	EventCommentCreated EventType = "comment_created"
)

// Event is broadcast to all connected WebSocket clients.
type Event struct {
	Type    EventType `json:"type"`
	Payload any       `json:"payload"`
	By      int64     `json:"-"` // acting user ID; hub skips clients matching this
}

// ── Client ──

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID int64
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// ── Hub ──

type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's event loop. Must be called as a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		}
	}
}

// Broadcast enqueues a message to every connected client except the one
// whose userID matches event.By (the acting user).
func (h *Hub) Broadcast(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.userID != 0 && client.userID == event.By {
			continue
		}
		select {
		case client.send <- data:
		default:
			// Slow client — message dropped.
		}
	}
}

// ── HTTP handler (method on *Handler, lives here for cohesion) ──

// ServeWs handles the GET /api/ws WebSocket upgrade.
// Auth uses ?pat= query param (browser WebSocket API cannot set custom headers).
func (h *Handler) ServeWs(w http.ResponseWriter, r *http.Request) {
	pat := r.URL.Query().Get("pat")
	if pat == "" {
		http.Error(w, `{"error":"missing pat"}`, http.StatusUnauthorized)
		return
	}
	user, err := h.store.GetUserByPAT(pat)
	if err != nil {
		http.Error(w, `{"error":"invalid pat"}`, http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		hub:    h.hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: user.ID,
	}
	h.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
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

// ── Diff helpers ──

// diffIssue returns only the fields that changed between two Issue snapshots.
func diffIssue(before, after *Issue) map[string]any {
	changed := map[string]any{}
	if before.Title != after.Title {
		changed["title"] = diff("before", before.Title, after.Title)
	}
	if before.Description != after.Description {
		changed["description"] = diff("before", before.Description, after.Description)
	}
	if before.Type != after.Type {
		changed["type"] = diff("before", before.Type, after.Type)
	}
	if before.State != after.State {
		changed["state"] = diff("before", before.State, after.State)
	}
	if before.Assignee != after.Assignee {
		changed["assignee"] = diff("before", nullableInt(before.Assignee), nullableInt(after.Assignee))
	}
	if before.Priority != after.Priority {
		changed["priority"] = diff("before", before.Priority, after.Priority)
	}
	if before.ParentID != after.ParentID {
		changed["parent_id"] = diff("before", nullableInt(before.ParentID), nullableInt(after.ParentID))
	}
	return changed
}

// diffProject returns only the fields that changed between two Project snapshots.
func diffProject(before, after *Project) map[string]any {
	changed := map[string]any{}
	if before.Name != after.Name {
		changed["name"] = diff("before", before.Name, after.Name)
	}
	if before.Slug != after.Slug {
		changed["slug"] = diff("before", before.Slug, after.Slug)
	}
	if before.Description != after.Description {
		changed["description"] = diff("before", before.Description, after.Description)
	}
	return changed
}

// diff builds a {"before": v1, "after": v2} map for one changed field.
func diff[T any](_ string, before, after T) map[string]T {
	return map[string]T{"before": before, "after": after}
}
