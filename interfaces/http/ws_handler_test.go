package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	gorillaws "github.com/gorilla/websocket"
	"github.com/waiter/back/infrastructure/auth"
	ws "github.com/waiter/back/infrastructure/websocket"
)

var wsTestSecret = []byte("ws-test-secret-32bytes-enough!!!")
var wsAdminSecret = []byte("ws-admin-secret-32bytes-enough!!")

func makeWSToken(restaurantID string, expOffset time.Duration) string {
	claims := auth.SessionClaims{
		TableID:      "t1",
		RestaurantID: restaurantID,
		TableNumber:  1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expOffset)),
		},
	}
	token, _ := auth.SignSession(claims, wsTestSecret)
	return token
}

func makeAdminWSToken(restaurantID string, expOffset time.Duration) string {
	claims := auth.AdminClaims{
		AdminID:      "admin-1",
		RestaurantID: restaurantID,
		Role:         "owner",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expOffset)),
		},
	}
	token, _ := auth.SignAdminSession(claims, wsAdminSecret)
	return token
}

func buildWSRouter(secret []byte) (*WSHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	hub := ws.NewHub(nil)
	handler := NewWSHandler(hub, secret, wsAdminSecret)
	r := gin.New()
	r.GET("/ws/:restaurantId", handler.Connect)
	return handler, r
}

func TestNewWSHandler(t *testing.T) {
	hub := ws.NewHub(nil)
	handler := NewWSHandler(hub, wsTestSecret, wsAdminSecret)
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
	if handler.hub != hub {
		t.Error("expected hub to be set")
	}
}

func TestWSHandler_Connect_Success(t *testing.T) {
	_, r := buildWSRouter(wsTestSecret)
	server := httptest.NewServer(r)
	defer server.Close()

	token := makeWSToken("r1", 30*time.Minute)
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/r1?token=" + token
	conn, _, err := gorillaws.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("expected successful WS connection: %v", err)
	}
	conn.Close()
	time.Sleep(50 * time.Millisecond)
}

func TestWSHandler_Connect_NoToken(t *testing.T) {
	_, r := buildWSRouter(wsTestSecret)
	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/r1"
	_, resp, err := gorillaws.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected connection to be rejected without token")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %v", resp)
	}
}

func TestWSHandler_Connect_InvalidToken(t *testing.T) {
	_, r := buildWSRouter(wsTestSecret)
	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/r1?token=not.a.valid.jwt"
	_, resp, err := gorillaws.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected connection to be rejected with invalid token")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %v", resp)
	}
}

func TestWSHandler_Connect_WrongRestaurant(t *testing.T) {
	_, r := buildWSRouter(wsTestSecret)
	server := httptest.NewServer(r)
	defer server.Close()

	// Token for r2, but connecting to r1
	token := makeWSToken("r2", 30*time.Minute)
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/r1?token=" + token
	_, resp, err := gorillaws.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected connection to be rejected for wrong restaurant")
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %v", resp)
	}
}

func TestWSHandler_Connect_AdminToken_Success(t *testing.T) {
	_, r := buildWSRouter(wsTestSecret)
	server := httptest.NewServer(r)
	defer server.Close()

	token := makeAdminWSToken("r1", 30*time.Minute)
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/r1?token=" + token
	conn, _, err := gorillaws.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("expected admin WS connection to succeed: %v", err)
	}
	conn.Close()
	time.Sleep(50 * time.Millisecond)
}

func TestWSHandler_Connect_AdminToken_WrongRestaurant(t *testing.T) {
	_, r := buildWSRouter(wsTestSecret)
	server := httptest.NewServer(r)
	defer server.Close()

	// Admin token for r2, connecting to r1
	token := makeAdminWSToken("r2", 30*time.Minute)
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/r1?token=" + token
	_, resp, err := gorillaws.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected connection to be rejected for wrong restaurant")
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %v", resp)
	}
}
