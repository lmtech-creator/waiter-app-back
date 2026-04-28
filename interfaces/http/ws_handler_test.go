package http

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	gorillaws "github.com/gorilla/websocket"
	ws "github.com/waiter/back/infrastructure/websocket"
)

func TestNewWSHandler(t *testing.T) {
	hub := ws.NewHub()
	handler := NewWSHandler(hub)
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
	if handler.hub != hub {
		t.Error("expected hub to be set")
	}
}

func TestWSHandler_Connect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := ws.NewHub()
	wsHandler := NewWSHandler(hub)

	r := gin.New()
	r.GET("/ws/:restaurantId", wsHandler.Connect)

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/r1"
	conn, _, err := gorillaws.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect to websocket: %v", err)
	}
	conn.Close()
	time.Sleep(50 * time.Millisecond)
}
