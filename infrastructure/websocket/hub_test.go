package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("expected non-nil hub")
	}
	if hub.rooms == nil {
		t.Fatal("expected initialized rooms map")
	}
}

func TestHub_Notify_NoClients(t *testing.T) {
	hub := NewHub()
	// Should not panic with no clients
	hub.Notify("r1", map[string]string{"event": "test"})
}

func TestHub_Notify_InvalidJSON(t *testing.T) {
	hub := NewHub()
	// Should not panic with unmarshalable data
	hub.Notify("r1", make(chan int))
}

func TestHub_SubscribeAndNotify(t *testing.T) {
	hub := NewHub()

	// Create test server with the hub
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub.Subscribe(w, r, "rest1")
	}))
	defer server.Close()

	// Connect WebSocket client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Give time for subscription to register
	time.Sleep(50 * time.Millisecond)

	// Send notification
	hub.Notify("rest1", map[string]string{"event": "new_request"})

	// Read message
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var data map[string]string
	if err := json.Unmarshal(message, &data); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}
	if data["event"] != "new_request" {
		t.Errorf("expected event 'new_request', got '%s'", data["event"])
	}
}

func TestHub_MultipleClients(t *testing.T) {
	hub := NewHub()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub.Subscribe(w, r, "rest1")
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect 2 clients
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect client 1: %v", err)
	}
	defer conn1.Close()

	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect client 2: %v", err)
	}
	defer conn2.Close()

	time.Sleep(50 * time.Millisecond)

	hub.Notify("rest1", map[string]string{"event": "test"})

	for i, conn := range []*websocket.Conn{conn1, conn2} {
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, msg, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("client %d failed to read: %v", i+1, err)
		}
		var data map[string]string
		if err := json.Unmarshal(msg, &data); err != nil {
			t.Fatalf("client %d unmarshal error: %v", i+1, err)
		}
		if data["event"] != "test" {
			t.Errorf("client %d expected event 'test', got '%s'", i+1, data["event"])
		}
	}
}

func TestHub_ClientDisconnect(t *testing.T) {
	hub := NewHub()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub.Subscribe(w, r, "rest1")
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Close client connection
	conn.Close()

	// Give time for cleanup
	time.Sleep(100 * time.Millisecond)

	hub.mu.RLock()
	clients := hub.rooms["rest1"]
	hub.mu.RUnlock()

	if len(clients) != 0 {
		t.Errorf("expected 0 clients after disconnect, got %d", len(clients))
	}
}

func TestHub_DifferentRooms(t *testing.T) {
	hub := NewHub()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		room := r.URL.Query().Get("room")
		hub.Subscribe(w, r, room)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	conn1, _, err := websocket.DefaultDialer.Dial(wsURL+"?room=r1", nil)
	if err != nil {
		t.Fatalf("failed to connect client 1: %v", err)
	}
	defer conn1.Close()

	conn2, _, err := websocket.DefaultDialer.Dial(wsURL+"?room=r2", nil)
	if err != nil {
		t.Fatalf("failed to connect client 2: %v", err)
	}
	defer conn2.Close()

	time.Sleep(50 * time.Millisecond)

	// Notify only room r1
	hub.Notify("r1", map[string]string{"event": "only_r1"})

	conn1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn1.ReadMessage()
	if err != nil {
		t.Fatalf("client 1 failed to read: %v", err)
	}
	var data map[string]string
	if err := json.Unmarshal(msg, &data); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if data["event"] != "only_r1" {
		t.Errorf("expected 'only_r1', got '%s'", data["event"])
	}

	// Client 2 should NOT receive the message (timeout expected)
	conn2.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, _, err = conn2.ReadMessage()
	if err == nil {
		t.Error("client 2 should not have received a message for room r1")
	}
}
