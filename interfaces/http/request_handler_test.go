package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/application/usecase"
	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/infrastructure/auth"
	"github.com/waiter/back/interfaces/http/dto"
	mw "github.com/waiter/back/interfaces/http/middleware"
	"github.com/waiter/back/mocks"
)

func setupRequestRouter() (*RequestHandler, *mocks.RequestRepo, *mocks.TableRepo, *mocks.Notifier) {
	requestRepo := mocks.NewRequestRepo()
	tableRepo := mocks.NewTableRepo()
	notifier := mocks.NewNotifier()
	uc := usecase.NewRequestUseCase(requestRepo, tableRepo, notifier)
	handler := NewRequestHandler(uc)
	return handler, requestRepo, tableRepo, notifier
}

func injectClaims(c *gin.Context, tableID, restaurantID string, tableNumber int) {
	c.Set(mw.ClaimsKey, &auth.SessionClaims{
		TableID:      tableID,
		RestaurantID: restaurantID,
		TableNumber:  tableNumber,
	})
}

func TestRequestHandler_Create_Success(t *testing.T) {
	handler, _, _, _ := setupRequestRouter()

	body, _ := json.Marshal(map[string]any{"type": "CALL_WAITER"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/requests", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	injectClaims(c, "t1", "r1", 1)

	handler.Create(c)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	var resp dto.RequestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Type != entity.CallWaiter {
		t.Errorf("expected type CALL_WAITER, got %s", resp.Type)
	}
}

func TestRequestHandler_Create_BadJSON(t *testing.T) {
	handler, _, _, _ := setupRequestRouter()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/requests", bytes.NewReader([]byte(`{bad`)))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRequestHandler_Create_NoClaims(t *testing.T) {
	handler, _, _, _ := setupRequestRouter()

	body, _ := json.Marshal(map[string]any{"type": "CALL_WAITER"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/requests", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	// No claims injected — simulates missing auth middleware

	handler.Create(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequestHandler_Create_UseCaseError(t *testing.T) {
	handler, requestRepo, _, _ := setupRequestRouter()
	requestRepo.CreateFn = func(r *entity.Request) error {
		return fmt.Errorf("db error")
	}

	body, _ := json.Marshal(map[string]any{"type": "CALL_WAITER"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/requests", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	injectClaims(c, "t1", "r1", 1)

	handler.Create(c)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
}

func TestRequestHandler_GetActive_Success(t *testing.T) {
	handler, requestRepo, _, _ := setupRequestRouter()
	requestRepo.Requests["r1"] = &entity.Request{ID: "r1", TableID: "t1", Status: entity.Pending}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/rest1/requests/active", nil)
	c.Params = gin.Params{{Key: "restaurantId", Value: "rest1"}}

	handler.GetActive(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequestHandler_GetActive_RepoError(t *testing.T) {
	handler, requestRepo, _, _ := setupRequestRouter()
	requestRepo.FindActiveFn = func(restaurantID string) ([]entity.Request, error) {
		return nil, fmt.Errorf("db error")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/rest1/requests/active", nil)
	c.Params = gin.Params{{Key: "restaurantId", Value: "rest1"}}

	handler.GetActive(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestRequestHandler_Complete_Success(t *testing.T) {
	handler, requestRepo, tableRepo, _ := setupRequestRouter()
	requestRepo.Requests["req1"] = &entity.Request{ID: "req1", TableID: "t1", Status: entity.Pending}
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", RestaurantID: "r1"}

	body, _ := json.Marshal(map[string]any{"status": "DONE"})
	w := httptest.NewRecorder()
	r := gin.New()
	r.PATCH("/api/v1/requests/:requestId", handler.Complete)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/requests/req1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRequestHandler_Complete_BadJSON(t *testing.T) {
	handler, _, _, _ := setupRequestRouter()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/requests/req1", bytes.NewReader([]byte(`{bad`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "requestId", Value: "req1"}}

	handler.Complete(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRequestHandler_Complete_InvalidStatus(t *testing.T) {
	handler, _, _, _ := setupRequestRouter()

	body, _ := json.Marshal(map[string]any{"status": "PENDING"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/requests/req1", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "requestId", Value: "req1"}}

	handler.Complete(c)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
}

func TestRequestHandler_Complete_UseCaseError(t *testing.T) {
	handler, _, _, _ := setupRequestRouter()

	body, _ := json.Marshal(map[string]any{"status": "DONE"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/requests/nonexistent", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "requestId", Value: "nonexistent"}}

	handler.Complete(c)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
}

func TestRequestHandler_GetTableStatus_Success(t *testing.T) {
	handler, requestRepo, _, _ := setupRequestRouter()
	requestRepo.Requests["r1"] = &entity.Request{ID: "r1", TableID: "t1", Status: entity.Pending}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/tables/t1/status", nil)
	c.Params = gin.Params{{Key: "tableId", Value: "t1"}}

	handler.GetTableStatus(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequestHandler_GetTableStatus_RepoError(t *testing.T) {
	handler, requestRepo, _, _ := setupRequestRouter()
	requestRepo.FindByTableFn = func(tableID string) ([]entity.Request, error) {
		return nil, fmt.Errorf("db error")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/tables/t1/status", nil)
	c.Params = gin.Params{{Key: "tableId", Value: "t1"}}

	handler.GetTableStatus(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestRequestHandler_Create_CooldownRejected(t *testing.T) {
	handler, requestRepo, _, _ := setupRequestRouter()
	requestRepo.FindLastCreatedByTableFn = func(tableID string) (*entity.Request, error) {
		return &entity.Request{
			ID:        "prev",
			TableID:   tableID,
			CreatedAt: time.Now().Add(-5 * time.Second),
		}, nil
	}

	body, _ := json.Marshal(map[string]any{"type": "CALL_WAITER"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/requests", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	injectClaims(c, "t1", "r1", 1)

	handler.Create(c)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d: %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header to be set")
	}
}

func TestRequestHandler_Create_InvalidType(t *testing.T) {
	handler, _, _, _ := setupRequestRouter()

	body, _ := json.Marshal(map[string]any{"type": "INVALID_TYPE"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/requests", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	injectClaims(c, "t1", "r1", 1)

	handler.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
