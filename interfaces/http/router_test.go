package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/waiter/back/application/usecase"
	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/infrastructure/auth"
	ws "github.com/waiter/back/infrastructure/websocket"
	"github.com/waiter/back/mocks"
)

var routerAdminSecret = []byte("admin-test-secret-32bytes-enough")

func makeRouterAdminToken(t *testing.T) string {
	t.Helper()
	claims := auth.AdminClaims{
		AdminID:      "admin-1",
		RestaurantID: "r1",
		Role:         string(entity.RoleSuperAdmin),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token, err := auth.SignAdminSession(claims, routerAdminSecret)
	if err != nil {
		t.Fatalf("could not sign admin token: %v", err)
	}
	return token
}

func buildTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	requestRepo := mocks.NewRequestRepo()
	tableRepo := mocks.NewTableRepo()
	notifier := mocks.NewNotifier()
	feedbackRepo := mocks.NewFeedbackRepo()
	restaurantRepo := mocks.NewRestaurantRepo()
	adminRepo := mocks.NewAdminRepo()

	requestUC := usecase.NewRequestUseCase(requestRepo, tableRepo, notifier)
	feedbackUC := usecase.NewFeedbackUseCase(feedbackRepo, tableRepo)
	restaurantUC := usecase.NewRestaurantUseCase(restaurantRepo, tableRepo)
	adminUC := usecase.NewAdminUseCase(adminRepo, []byte("admin-test-secret-32bytes-enough"))

	requestHandler := NewRequestHandler(requestUC)
	feedbackHandler := NewFeedbackHandler(feedbackUC)
	restaurantHandler := NewRestaurantHandler(restaurantUC)
	sessionHandler := NewSessionHandler(tableRepo, []byte("test-secret-that-is-long-enough!!"))
	adminHandler := NewAdminHandler(adminUC)
	hub := ws.NewHub(nil)

	return SetupRouter(requestHandler, feedbackHandler, restaurantHandler, sessionHandler, adminHandler, hub, []byte("test-secret-that-is-long-enough!!"), []byte("admin-test-secret-32bytes-enough"))
}

func TestSetupRouter_HealthCheck(t *testing.T) {
	router := buildTestRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestSetupRouter_RoutesExist(t *testing.T) {
	router := buildTestRouter(t)

	routes := router.Routes()
	expectedPaths := map[string]string{
		"POST:/api/v1/session":                                         "",
		"POST:/api/v1/requests":                                        "",
		"GET:/api/v1/tables/:tableId/status":                           "",
		"POST:/api/v1/feedback":                                        "",
		"GET:/api/v1/restaurants/:restaurantId/requests/active":        "",
		"PATCH:/api/v1/requests/:requestId":                            "",
		"POST:/api/v1/restaurants":                                     "",
		"GET:/api/v1/restaurants":                                      "",
		"GET:/api/v1/restaurants/:restaurantId":                        "",
		"POST:/api/v1/restaurants/:restaurantId/tables":                "",
		"GET:/api/v1/restaurants/:restaurantId/tables":                 "",
		"POST:/api/v1/restaurants/:restaurantId/tables/:id/inactivate": "",
		"POST:/api/v1/restaurants/:restaurantId/tables/:id/activate":   "",
		"POST:/api/v1/admin/tables/:id/regenerate-qr":                  "",
		"POST:/api/v1/admin/login":                                     "",
		"POST:/api/v1/admin/users":                                     "",
		"GET:/api/v1/admin/users":                                      "",
		"DELETE:/api/v1/admin/users/:id":                               "",
		"GET:/api/v1/ws/:restaurantId":                                 "",
		"GET:/health":                                                  "",
	}

	for _, r := range routes {
		key := r.Method + ":" + r.Path
		delete(expectedPaths, key)
	}

	for path := range expectedPaths {
		t.Errorf("expected route not found: %s", path)
	}
}

func TestSetupRouter_IntegrationCreateRestaurant(t *testing.T) {
	router := buildTestRouter(t)

	w := httptest.NewRecorder()
	body := `{"name":"Integration Test Restaurant"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/restaurants", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeRouterAdminToken(t))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}
