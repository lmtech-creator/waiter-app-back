package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/waiter/back/application/usecase"
	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/mocks"
	"golang.org/x/crypto/bcrypt"
)

var adminHandlerSecret = []byte("admin-handler-secret-32bytes-ok!")

func hashForTest(t *testing.T, pw string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	return string(h)
}

func buildAdminRouter(t *testing.T) (*gin.Engine, *mocks.AdminRepo) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	repo := mocks.NewAdminRepo()
	uc := usecase.NewAdminUseCase(repo, adminHandlerSecret)
	h := NewAdminHandler(uc)
	r := gin.New()
	r.POST("/api/v1/admin/login", h.Login)
	return r, repo
}

func TestAdminHandler_Login_Success(t *testing.T) {
	r, repo := buildAdminRouter(t)
	rid := uuid.New().String()
	repo.Admins["a1"] = &entity.AdminUser{
		ID:           "a1",
		RestaurantID: rid,
		Username:     "admin",
		PasswordHash: hashForTest(t, "secret123"),
	}

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "secret123"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["token"] == "" {
		t.Error("expected non-empty token in response")
	}
}

func TestAdminHandler_Login_WrongPassword(t *testing.T) {
	r, repo := buildAdminRouter(t)
	repo.Admins["a1"] = &entity.AdminUser{
		ID:           "a1",
		Username:     "admin",
		PasswordHash: hashForTest(t, "correct"),
	}

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "wrong"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAdminHandler_Login_EmptyBody(t *testing.T) {
	r, _ := buildAdminRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/login", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAdminHandler_Login_InvalidJSON(t *testing.T) {
	r, _ := buildAdminRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/login", bytes.NewBufferString(`not-json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAdminHandler_Login_UnknownUser(t *testing.T) {
	r, _ := buildAdminRouter(t)

	body, _ := json.Marshal(map[string]string{"username": "ghost", "password": "pass"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
