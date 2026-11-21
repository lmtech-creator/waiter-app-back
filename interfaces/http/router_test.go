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
	return makeRouterToken(t, entity.RoleSuperAdmin, "r1")
}

func makeRouterToken(t *testing.T, role entity.AdminRole, restaurantID string) string {
	t.Helper()
	claims := auth.AdminClaims{
		AdminID:      "admin-1",
		RestaurantID: restaurantID,
		Role:         string(role),
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
	restaurantRepo.Restaurants["r1"] = &entity.Restaurant{ID: "r1", Name: "República"}
	adminRepo := mocks.NewAdminRepo()

	planRepo := mocks.NewPlanRepo()
	planRepo.Plans["plan-free"] = &entity.Plan{ID: "plan-free", Name: "free", Description: "Plan gratuito"}

	requestUC := usecase.NewRequestUseCase(requestRepo, tableRepo, notifier)
	feedbackUC := usecase.NewFeedbackUseCase(feedbackRepo, tableRepo)
	restaurantUC := usecase.NewRestaurantUseCase(restaurantRepo, tableRepo, planRepo)
	adminUC := usecase.NewAdminUseCase(adminRepo, []byte("admin-test-secret-32bytes-enough"))
	statsUC := usecase.NewStatsUseCase(requestRepo, feedbackRepo)
	planUC := usecase.NewPlanUseCase(planRepo)

	requestHandler := NewRequestHandler(requestUC)
	feedbackHandler := NewFeedbackHandler(feedbackUC)
	restaurantHandler := NewRestaurantHandler(restaurantUC)
	sessionHandler := NewSessionHandler(tableRepo, restaurantRepo, []byte("test-secret-that-is-long-enough!!"))
	adminHandler := NewAdminHandler(adminUC)
	statsHandler := NewStatsHandler(statsUC)
	planHandler := NewPlanHandler(planUC)
	hub := ws.NewHub(nil)

	return SetupRouter(requestHandler, feedbackHandler, restaurantHandler, sessionHandler, adminHandler, statsHandler, planHandler, hub, []byte("test-secret-that-is-long-enough!!"), []byte("admin-test-secret-32bytes-enough"))
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
		"GET:/api/v1/requests/:requestId":                              "",
		"GET:/api/v1/tables/:tableId/status":                           "",
		"POST:/api/v1/feedback":                                        "",
		"GET:/api/v1/restaurants/:restaurantId/requests/active":        "",
		"GET:/api/v1/restaurants/:restaurantId/feedback":               "",
		"PATCH:/api/v1/requests/:requestId":                            "",
		"POST:/api/v1/restaurants":                                     "",
		"GET:/api/v1/restaurants":                                      "",
		"GET:/api/v1/restaurants/:restaurantId":                        "",
		"PATCH:/api/v1/restaurants/:restaurantId":                      "",
		"POST:/api/v1/restaurants/:restaurantId/tables":                "",
		"GET:/api/v1/restaurants/:restaurantId/tables":                 "",
		"POST:/api/v1/restaurants/:restaurantId/tables/:id/inactivate": "",
		"POST:/api/v1/restaurants/:restaurantId/tables/:id/activate":   "",
		"DELETE:/api/v1/restaurants/:restaurantId/tables/:id":          "",
		"POST:/api/v1/admin/tables/:id/regenerate-qr":                  "",
		"POST:/api/v1/admin/login":                                     "",
		"POST:/api/v1/admin/users":                                     "",
		"GET:/api/v1/admin/users":                                      "",
		"PATCH:/api/v1/admin/users/:id":                                "",
		"POST:/api/v1/admin/users/:id/reset-password":                  "",
		"DELETE:/api/v1/admin/users/:id":                               "",
		"GET:/api/v1/plans":                                            "",
		"POST:/api/v1/plans":                                           "",
		"GET:/api/v1/plans/:planId":                                    "",
		"PATCH:/api/v1/plans/:planId":                                  "",
		"DELETE:/api/v1/plans/:planId":                                 "",
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

func TestSetupRouter_EmployeeReadAccess(t *testing.T) {
	router := buildTestRouter(t)
	employee := makeRouterToken(t, entity.RoleEmployee, "r1")

	get := func(path string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+employee)
		router.ServeHTTP(w, req)
		return w.Code
	}

	// Waiters must be able to see their own restaurant, tables, live requests
	// and stats.
	if code := get("/api/v1/restaurants/r1"); code != http.StatusOK {
		t.Errorf("employee GET own restaurant: expected 200, got %d", code)
	}
	if code := get("/api/v1/restaurants/r2"); code != http.StatusForbidden {
		t.Errorf("employee GET other restaurant: expected 403, got %d", code)
	}
	if code := get("/api/v1/restaurants/r1/tables"); code != http.StatusOK {
		t.Errorf("employee GET tables: expected 200, got %d", code)
	}
	if code := get("/api/v1/restaurants/r1/requests/active"); code != http.StatusOK {
		t.Errorf("employee GET active requests: expected 200, got %d", code)
	}
	if code := get("/api/v1/restaurants/r1/stats?since=2026-01-01T00:00:00Z&until=2026-01-02T00:00:00Z"); code != http.StatusOK {
		t.Errorf("employee GET stats: expected 200, got %d", code)
	}

	// Owner-only surfaces stay forbidden.
	if code := get("/api/v1/restaurants/r1/feedback"); code != http.StatusForbidden {
		t.Errorf("employee GET feedback: expected 403, got %d", code)
	}
	if code := get("/api/v1/restaurants/r2/tables"); code != http.StatusForbidden {
		t.Errorf("employee GET other restaurant tables: expected 403, got %d", code)
	}

	// Mutations stay owner-only.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/restaurants/r1/tables", bytes.NewBufferString(`{"number":1}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+employee)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("employee POST tables: expected 403, got %d", w.Code)
	}
}

func TestSetupRouter_MenuUpdatePermissions(t *testing.T) {
	router := buildTestRouter(t)
	owner := makeRouterToken(t, entity.RoleOwner, "r1")
	employee := makeRouterToken(t, entity.RoleEmployee, "r1")

	patch := func(path, token string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPatch, path, bytes.NewBufferString(`{"menu_url":"https://example.com/menu"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
		return w.Code
	}

	if code := patch("/api/v1/restaurants/r1", owner); code != http.StatusOK {
		t.Errorf("owner PATCH own restaurant: expected 200, got %d", code)
	}
	if code := patch("/api/v1/restaurants/r2", owner); code != http.StatusForbidden {
		t.Errorf("owner PATCH other restaurant: expected 403, got %d", code)
	}
	if code := patch("/api/v1/restaurants/r1", employee); code != http.StatusForbidden {
		t.Errorf("employee PATCH restaurant: expected 403, got %d", code)
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
