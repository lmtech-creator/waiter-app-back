package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/infrastructure/auth"
	"github.com/waiter/back/interfaces/http/dto"
	"github.com/waiter/back/mocks"
)

func setupSessionHandler() (*SessionHandler, *mocks.TableRepo) {
	tableRepo := mocks.NewTableRepo()
	secret := []byte("test-secret-32bytes-long-enough!!")
	handler := NewSessionHandler(tableRepo, secret)
	return handler, tableRepo
}

func TestSessionHandler_Create_Success(t *testing.T) {
	handler, tableRepo := setupSessionHandler()
	tableRepo.Tables["t1"] = &entity.Table{
		ID:           "t1",
		Number:       5,
		RestaurantID: "r1",
		QRCode:       "ABCDEF1234",
		IsActive:     true,
	}

	body, _ := json.Marshal(map[string]any{"qr_code": "ABCDEF1234"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp dto.SessionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.SessionToken == "" {
		t.Error("expected non-empty session token")
	}
	if resp.Table.Number != 5 {
		t.Errorf("expected table number 5, got %d", resp.Table.Number)
	}
}

func TestSessionHandler_Create_BadJSON(t *testing.T) {
	handler, _ := setupSessionHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewReader([]byte(`{invalid`)))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSessionHandler_Create_EmptyQRCode(t *testing.T) {
	handler, _ := setupSessionHandler()

	body, _ := json.Marshal(map[string]any{"qr_code": ""})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSessionHandler_Create_QRNotFound(t *testing.T) {
	handler, tableRepo := setupSessionHandler()
	tableRepo.FindByQRCodeFn = func(qrCode string) (*entity.Table, error) {
		return nil, fmt.Errorf("not found")
	}

	body, _ := json.Marshal(map[string]any{"qr_code": "NOTEXIST99"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestSessionHandler_Create_TableInactive(t *testing.T) {
	handler, tableRepo := setupSessionHandler()
	tableRepo.Tables["t2"] = &entity.Table{
		ID:       "t2",
		Number:   3,
		QRCode:   "INACTIVE01",
		IsActive: false,
	}

	body, _ := json.Marshal(map[string]any{"qr_code": "INACTIVE01"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestSessionHandler_Create_SignError(t *testing.T) {
	tableRepo := mocks.NewTableRepo()
	tableRepo.Tables["t1"] = &entity.Table{
		ID:       "t1",
		Number:   1,
		QRCode:   "SIGNERROR1",
		IsActive: true,
	}
	handler := &SessionHandler{
		tableRepo: tableRepo,
		secret:    []byte("test"),
		signSession: func(claims auth.SessionClaims, secret []byte) (string, error) {
			return "", fmt.Errorf("sign error")
		},
	}

	body, _ := json.Marshal(map[string]any{"qr_code": "SIGNERROR1"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
