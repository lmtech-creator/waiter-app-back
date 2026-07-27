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
	"github.com/waiter/back/infrastructure/auth"
	"github.com/waiter/back/interfaces/http/dto"
	mw "github.com/waiter/back/interfaces/http/middleware"
	"github.com/waiter/back/mocks"
	"golang.org/x/crypto/bcrypt"
)

func strPtr(s string) *string { return &s }

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
		RestaurantID: &rid,
		Username:     "admin",
		PasswordHash: hashForTest(t, "secret123"),
		Role:         entity.RoleOwner,
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
		Role:         entity.RoleOwner,
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

// ─── UpdateUser ───────────────────────────────────────────────────────────────

func injectAdminClaims(c *gin.Context, adminID, restaurantID, role string) {
	c.Set(mw.AdminClaimsKey, &auth.AdminClaims{
		AdminID:      adminID,
		RestaurantID: restaurantID,
		Role:         role,
	})
}

func TestAdminHandler_UpdateUser_Success(t *testing.T) {
	repo := mocks.NewAdminRepo()
	repo.Admins["u1"] = &entity.AdminUser{
		ID: "u1", Username: "old", Role: entity.RoleEmployee, RestaurantID: strPtr("r1"),
	}
	uc := usecase.NewAdminUseCase(repo, adminHandlerSecret)
	h := NewAdminHandler(uc)

	body, _ := json.Marshal(map[string]any{"username": "new-name"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/u1", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "u1"}}
	injectAdminClaims(c, "admin1", "", "superadmin")

	h.UpdateUser(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp dto.AdminUserResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Username != "new-name" {
		t.Errorf("expected username 'new-name', got '%s'", resp.Username)
	}
}

func TestAdminHandler_UpdateUser_Forbidden(t *testing.T) {
	repo := mocks.NewAdminRepo()
	repo.Admins["u1"] = &entity.AdminUser{
		ID: "u1", Username: "owner2", Role: entity.RoleOwner, RestaurantID: strPtr("r1"),
	}
	uc := usecase.NewAdminUseCase(repo, adminHandlerSecret)
	h := NewAdminHandler(uc)

	body, _ := json.Marshal(map[string]any{"username": "hacker"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/u1", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "u1"}}
	injectAdminClaims(c, "owner1", "r1", "owner")

	h.UpdateUser(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestAdminHandler_UpdateUser_NotFound(t *testing.T) {
	repo := mocks.NewAdminRepo()
	uc := usecase.NewAdminUseCase(repo, adminHandlerSecret)
	h := NewAdminHandler(uc)

	body, _ := json.Marshal(map[string]any{"username": "any"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/nox", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "nox"}}
	injectAdminClaims(c, "admin1", "", "superadmin")

	h.UpdateUser(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestAdminHandler_UpdateUser_BadJSON(t *testing.T) {
	repo := mocks.NewAdminRepo()
	uc := usecase.NewAdminUseCase(repo, adminHandlerSecret)
	h := NewAdminHandler(uc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/u1", bytes.NewReader([]byte(`{bad`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "u1"}}
	injectAdminClaims(c, "admin1", "", "superadmin")

	h.UpdateUser(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ─── ResetPassword ────────────────────────────────────────────────────────────

func TestAdminHandler_ResetPassword_Success(t *testing.T) {
	repo := mocks.NewAdminRepo()
	repo.Admins["u1"] = &entity.AdminUser{
		ID: "u1", Username: "emp", Role: entity.RoleEmployee, RestaurantID: strPtr("r1"),
	}
	uc := usecase.NewAdminUseCase(repo, adminHandlerSecret)
	h := NewAdminHandler(uc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/u1/reset-password", nil)
	c.Params = gin.Params{{Key: "id", Value: "u1"}}
	injectAdminClaims(c, "admin1", "", "superadmin")

	h.ResetPassword(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp dto.ResetPasswordResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.NewPassword == "" {
		t.Error("expected non-empty new_password")
	}
}

func TestAdminHandler_ResetPassword_Forbidden(t *testing.T) {
	repo := mocks.NewAdminRepo()
	repo.Admins["u1"] = &entity.AdminUser{
		ID: "u1", Username: "other", Role: entity.RoleEmployee, RestaurantID: strPtr("r2"),
	}
	uc := usecase.NewAdminUseCase(repo, adminHandlerSecret)
	h := NewAdminHandler(uc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/u1/reset-password", nil)
	c.Params = gin.Params{{Key: "id", Value: "u1"}}
	injectAdminClaims(c, "owner1", "r1", "owner")

	h.ResetPassword(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestAdminHandler_ResetPassword_NotFound(t *testing.T) {
	repo := mocks.NewAdminRepo()
	uc := usecase.NewAdminUseCase(repo, adminHandlerSecret)
	h := NewAdminHandler(uc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/nox/reset-password", nil)
	c.Params = gin.Params{{Key: "id", Value: "nox"}}
	injectAdminClaims(c, "admin1", "", "superadmin")

	h.ResetPassword(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
