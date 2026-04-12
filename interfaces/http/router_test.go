package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/application/usecase"
	ws "github.com/waiter/back/infrastructure/websocket"
	"github.com/waiter/back/mocks"
)

func TestSetupRouter_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	requestRepo := mocks.NewRequestRepo()
	tableRepo := mocks.NewTableRepo()
	notifier := mocks.NewNotifier()
	feedbackRepo := mocks.NewFeedbackRepo()
	restaurantRepo := mocks.NewRestaurantRepo()

	requestUC := usecase.NewRequestUseCase(requestRepo, tableRepo, notifier)
	feedbackUC := usecase.NewFeedbackUseCase(feedbackRepo, tableRepo)
	restaurantUC := usecase.NewRestaurantUseCase(restaurantRepo, tableRepo)

	requestHandler := NewRequestHandler(requestUC)
	feedbackHandler := NewFeedbackHandler(feedbackUC)
	restaurantHandler := NewRestaurantHandler(restaurantUC)
	hub := ws.NewHub()

	router := SetupRouter(requestHandler, feedbackHandler, restaurantHandler, hub)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestSetupRouter_RoutesExist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	requestRepo := mocks.NewRequestRepo()
	tableRepo := mocks.NewTableRepo()
	notifier := mocks.NewNotifier()
	feedbackRepo := mocks.NewFeedbackRepo()
	restaurantRepo := mocks.NewRestaurantRepo()

	requestUC := usecase.NewRequestUseCase(requestRepo, tableRepo, notifier)
	feedbackUC := usecase.NewFeedbackUseCase(feedbackRepo, tableRepo)
	restaurantUC := usecase.NewRestaurantUseCase(restaurantRepo, tableRepo)

	requestHandler := NewRequestHandler(requestUC)
	feedbackHandler := NewFeedbackHandler(feedbackUC)
	restaurantHandler := NewRestaurantHandler(restaurantUC)
	hub := ws.NewHub()

	router := SetupRouter(requestHandler, feedbackHandler, restaurantHandler, hub)

	routes := router.Routes()
	expectedPaths := map[string]string{
		"POST:/api/v1/requests":                                 "",
		"GET:/api/v1/tables/:tableId/status":                    "",
		"POST:/api/v1/feedback":                                 "",
		"GET:/api/v1/restaurants/:restaurantId/requests/active": "",
		"PATCH:/api/v1/requests/:requestId":                     "",
		"POST:/api/v1/restaurants":                              "",
		"GET:/api/v1/restaurants/:restaurantId":                 "",
		"POST:/api/v1/restaurants/:restaurantId/tables":         "",
		"GET:/api/v1/restaurants/:restaurantId/tables":          "",
		"GET:/api/v1/ws/:restaurantId":                          "",
		"GET:/health":                                           "",
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
	gin.SetMode(gin.TestMode)

	requestRepo := mocks.NewRequestRepo()
	tableRepo := mocks.NewTableRepo()
	notifier := mocks.NewNotifier()
	feedbackRepo := mocks.NewFeedbackRepo()
	restaurantRepo := mocks.NewRestaurantRepo()

	requestUC := usecase.NewRequestUseCase(requestRepo, tableRepo, notifier)
	feedbackUC := usecase.NewFeedbackUseCase(feedbackRepo, tableRepo)
	restaurantUC := usecase.NewRestaurantUseCase(restaurantRepo, tableRepo)

	requestHandler := NewRequestHandler(requestUC)
	feedbackHandler := NewFeedbackHandler(feedbackUC)
	restaurantHandler := NewRestaurantHandler(restaurantUC)
	hub := ws.NewHub()

	router := SetupRouter(requestHandler, feedbackHandler, restaurantHandler, hub)

	w := httptest.NewRecorder()
	body := `{"name":"Integration Test Restaurant"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/restaurants", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}
