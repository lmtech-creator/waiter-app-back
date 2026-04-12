package http

import (
	"testing"

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
