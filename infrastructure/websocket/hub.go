package websocket

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu       sync.RWMutex
	rooms    map[string]map[*websocket.Conn]bool
	upgrader websocket.Upgrader
}

// NewHub creates a new Hub.
// allowedOrigins is an allowlist of Origin header values; if empty, all origins are accepted (dev mode).
func NewHub(allowedOrigins []string) *Hub {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		if o != "" {
			allowed[o] = true
		}
	}
	return &Hub{
		rooms: make(map[string]map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				if len(allowed) == 0 {
					return true // dev: accept all origins
				}
				return allowed[r.Header.Get("Origin")]
			},
		},
	}
}

// Subscribe upgrades an HTTP connection to WebSocket and registers it to a restaurant room.
func (h *Hub) Subscribe(w http.ResponseWriter, r *http.Request, restaurantID string) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}

	h.mu.Lock()
	if h.rooms[restaurantID] == nil {
		h.rooms[restaurantID] = make(map[*websocket.Conn]bool)
	}
	h.rooms[restaurantID][conn] = true
	h.mu.Unlock()

	slog.Info("websocket client connected", "restaurant_id", restaurantID)

	// Keep connection alive; remove on disconnect.
	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.rooms[restaurantID], conn)
			if len(h.rooms[restaurantID]) == 0 {
				delete(h.rooms, restaurantID)
			}
			h.mu.Unlock()
			conn.Close()
			slog.Info("websocket client disconnected", "restaurant_id", restaurantID)
		}()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

// Notify broadcasts an event to all WebSocket connections in a restaurant room.
func (h *Hub) Notify(restaurantID string, event any) {
	data, err := json.Marshal(event)
	if err != nil {
		slog.Error("failed to marshal ws event", "error", err)
		return
	}

	h.mu.RLock()
	clients := h.rooms[restaurantID]
	h.mu.RUnlock()

	for conn := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			slog.Warn("failed to send ws message", "error", err)
			conn.Close()
			h.mu.Lock()
			delete(h.rooms[restaurantID], conn)
			h.mu.Unlock()
		}
	}
}
