package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/application/usecase"
	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/interfaces/http/dto"
	"github.com/waiter/back/mocks"
)

func setupRestaurantRouter() (*RestaurantHandler, *mocks.RestaurantRepo, *mocks.TableRepo, *mocks.PlanRepo) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	planRepo := mocks.NewPlanRepo()
	planRepo.Plans["plan-free"] = &entity.Plan{ID: "plan-free", Name: "free", Description: "Plan gratuito"}
	uc := usecase.NewRestaurantUseCase(restaurantRepo, tableRepo, planRepo)
	handler := NewRestaurantHandler(uc)
	return handler, restaurantRepo, tableRepo, planRepo
}

func TestRestaurantHandler_Create_Success(t *testing.T) {
	handler, _, _, _ := setupRestaurantRouter()

	body, _ := json.Marshal(map[string]any{"name": "Burger Place"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/restaurants", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	var resp dto.RestaurantResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Name != "Burger Place" {
		t.Errorf("expected name 'Burger Place', got %s", resp.Name)
	}
	if resp.Plan != "free" || resp.PlanID != "plan-free" {
		t.Errorf("expected default plan free (plan-free), got plan=%q plan_id=%q", resp.Plan, resp.PlanID)
	}
	if resp.QrBannerText != "Escaneá y llamá al mozo" {
		t.Errorf("expected default qr_banner_text, got %q", resp.QrBannerText)
	}
	if len(resp.QrFooterItems) != 3 {
		t.Errorf("expected 3 default footer items, got %d", len(resp.QrFooterItems))
	}
}

func TestRestaurantHandler_Create_WithCustomQR(t *testing.T) {
	handler, _, _, _ := setupRestaurantRouter()

	body, _ := json.Marshal(map[string]any{
		"name":            "QR Place",
		"qr_banner_text":  "Escaneá!",
		"qr_footer_items": []string{"Llamar", "Cuenta"},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/restaurants", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	var resp dto.RestaurantResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.QrBannerText != "Escaneá!" {
		t.Errorf("expected 'Escaneá!', got %q", resp.QrBannerText)
	}
	if len(resp.QrFooterItems) != 2 {
		t.Fatalf("expected 2 footer items, got %d", len(resp.QrFooterItems))
	}
	if resp.QrFooterItems[0] != "Llamar" {
		t.Errorf("expected 'Llamar', got %q", resp.QrFooterItems[0])
	}
}

func TestRestaurantHandler_Create_BadJSON(t *testing.T) {
	handler, _, _, _ := setupRestaurantRouter()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/restaurants", bytes.NewReader([]byte(`{bad`)))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRestaurantHandler_Create_InvalidBannerText(t *testing.T) {
	handler, _, _, _ := setupRestaurantRouter()

	body, _ := json.Marshal(map[string]any{
		"name":           "Test",
		"qr_banner_text": string(make([]byte, 101)),
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/restaurants", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
}

func TestRestaurantHandler_Create_InvalidFooterItems(t *testing.T) {
	handler, _, _, _ := setupRestaurantRouter()

	body, _ := json.Marshal(map[string]any{
		"name":            "Test",
		"qr_footer_items": []string{"a", "b", "c", "d", "e", "f"},
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/restaurants", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
}

func TestRestaurantHandler_Create_UseCaseError(t *testing.T) {
	handler, restaurantRepo, _, _ := setupRestaurantRouter()
	restaurantRepo.CreateFn = func(r *entity.Restaurant) error {
		return fmt.Errorf("db error")
	}

	body, _ := json.Marshal(map[string]any{"name": "Test"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/restaurants", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
}

func TestRestaurantHandler_Get_Success(t *testing.T) {
	handler, restaurantRepo, _, _ := setupRestaurantRouter()
	restaurantRepo.Restaurants["r1"] = &entity.Restaurant{
		ID: "r1", Name: "Sushi Bar", PlanID: "plan-free", Plan: entity.Plan{ID: "plan-free", Name: "free"},
		QrBannerText: "Escaneá!", QrFooterItems: entity.JSONB{"Llamar", "Cuenta"},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/r1", nil)
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}}

	handler.Get(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp dto.RestaurantResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Name != "Sushi Bar" {
		t.Errorf("expected 'Sushi Bar', got %s", resp.Name)
	}
	if resp.QrBannerText != "Escaneá!" {
		t.Errorf("expected 'Escaneá!', got %q", resp.QrBannerText)
	}
	if len(resp.QrFooterItems) != 2 {
		t.Fatalf("expected 2 footer items, got %d", len(resp.QrFooterItems))
	}
}

func TestRestaurantHandler_Get_NotFound(t *testing.T) {
	handler, _, _, _ := setupRestaurantRouter()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/nonexistent", nil)
	c.Params = gin.Params{{Key: "restaurantId", Value: "nonexistent"}}

	handler.Get(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestRestaurantHandler_CreateTable_Success(t *testing.T) {
	handler, restaurantRepo, _, _ := setupRestaurantRouter()
	restaurantRepo.Restaurants["r1"] = &entity.Restaurant{ID: "r1", Name: "Test"}

	body, _ := json.Marshal(map[string]any{"number": 3})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/restaurants/r1/tables", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}}

	handler.CreateTable(c)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	var resp dto.TableResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Number != 3 {
		t.Errorf("expected number 3, got %d", resp.Number)
	}
}

func TestRestaurantHandler_CreateTable_BadJSON(t *testing.T) {
	handler, _, _, _ := setupRestaurantRouter()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/restaurants/r1/tables", bytes.NewReader([]byte(`{bad`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}}

	handler.CreateTable(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRestaurantHandler_CreateTable_RestaurantNotFound(t *testing.T) {
	handler, _, _, _ := setupRestaurantRouter()

	body, _ := json.Marshal(map[string]any{"number": 1})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/restaurants/nonexistent/tables", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "restaurantId", Value: "nonexistent"}}

	handler.CreateTable(c)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
}

func TestRestaurantHandler_GetTables_Success(t *testing.T) {
	handler, _, tableRepo, _ := setupRestaurantRouter()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", RestaurantID: "r1", Number: 1}
	tableRepo.Tables["t2"] = &entity.Table{ID: "t2", RestaurantID: "r1", Number: 2}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/r1/tables", nil)
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}}

	handler.GetTables(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp []dto.TableResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 tables, got %d", len(resp))
	}
}

func TestRestaurantHandler_GetTables_RepoError(t *testing.T) {
	handler, _, tableRepo, _ := setupRestaurantRouter()
	tableRepo.FindByRestaurantFn = func(restaurantID string) ([]entity.Table, error) {
		return nil, fmt.Errorf("db error")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/r1/tables", nil)
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}}

	handler.GetTables(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestRestaurantHandler_RegenerateQR_Success(t *testing.T) {
	handler, _, tableRepo, _ := setupRestaurantRouter()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", Number: 1, QRCode: "OLDCODE123"}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/tables/t1/regenerate-qr", nil)
	c.Params = gin.Params{{Key: "id", Value: "t1"}}

	handler.RegenerateQR(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["qr_code"] == "" {
		t.Error("expected non-empty qr_code in response")
	}
}

func TestRestaurantHandler_RegenerateQR_NotFound(t *testing.T) {
	handler, _, _, _ := setupRestaurantRouter()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/tables/nonexistent/regenerate-qr", nil)
	c.Params = gin.Params{{Key: "id", Value: "nonexistent"}}

	handler.RegenerateQR(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestRestaurantHandler_RegenerateQR_UpdateError(t *testing.T) {
	handler, _, tableRepo, _ := setupRestaurantRouter()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", QRCode: "OLD1234567"}
	tableRepo.UpdateQRCodeFn = func(id, qrCode string) error {
		return fmt.Errorf("db error")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/tables/t1/regenerate-qr", nil)
	c.Params = gin.Params{{Key: "id", Value: "t1"}}

	handler.RegenerateQR(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestRestaurantHandler_InactivateTable_Success(t *testing.T) {
	handler, _, tableRepo, _ := setupRestaurantRouter()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", RestaurantID: "r1", IsActive: true}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/restaurants/r1/tables/t1/inactivate", nil)
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}, {Key: "id", Value: "t1"}}

	handler.InactivateTable(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if tableRepo.Tables["t1"].IsActive {
		t.Error("expected table to be inactive")
	}
}

func TestRestaurantHandler_InactivateTable_NotFound(t *testing.T) {
	handler, _, _, _ := setupRestaurantRouter()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/restaurants/r1/tables/missing/inactivate", nil)
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}, {Key: "id", Value: "missing"}}

	handler.InactivateTable(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestRestaurantHandler_ActivateTable_Success(t *testing.T) {
	handler, _, tableRepo, _ := setupRestaurantRouter()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", RestaurantID: "r1", IsActive: false}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/restaurants/r1/tables/t1/activate", nil)
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}, {Key: "id", Value: "t1"}}

	handler.ActivateTable(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !tableRepo.Tables["t1"].IsActive {
		t.Error("expected table to be active")
	}
}

func TestRestaurantHandler_ActivateTable_Mismatch(t *testing.T) {
	handler, _, tableRepo, _ := setupRestaurantRouter()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", RestaurantID: "r2", IsActive: false}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/restaurants/r1/tables/t1/activate", nil)
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}, {Key: "id", Value: "t1"}}

	handler.ActivateTable(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRestaurantHandler_UpdateMenu_Success(t *testing.T) {
	handler, restaurantRepo, _, _ := setupRestaurantRouter()
	restaurantRepo.Restaurants["r1"] = &entity.Restaurant{ID: "r1", Name: "República"}

	body, _ := json.Marshal(map[string]any{"menu_url": "https://example.com/menu.pdf"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/restaurants/r1", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}}

	handler.Update(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp dto.RestaurantResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.MenuURL != "https://example.com/menu.pdf" {
		t.Errorf("expected updated menu_url, got %q", resp.MenuURL)
	}
	if restaurantRepo.Restaurants["r1"].MenuURL != "https://example.com/menu.pdf" {
		t.Error("expected menu_url to be persisted")
	}
}

func TestRestaurantHandler_UpdateMenu_InvalidURL(t *testing.T) {
	handler, restaurantRepo, _, _ := setupRestaurantRouter()
	restaurantRepo.Restaurants["r1"] = &entity.Restaurant{ID: "r1", Name: "República"}

	body, _ := json.Marshal(map[string]any{"menu_url": "javascript:alert(1)"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/restaurants/r1", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}}

	handler.Update(c)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
}

func TestRestaurantHandler_UpdateMenu_MissingField(t *testing.T) {
	handler, restaurantRepo, _, _ := setupRestaurantRouter()
	restaurantRepo.Restaurants["r1"] = &entity.Restaurant{ID: "r1", Name: "República"}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/restaurants/r1", bytes.NewReader([]byte(`{}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}}

	handler.Update(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRestaurantHandler_UpdateMenu_NotFound(t *testing.T) {
	handler, _, _, _ := setupRestaurantRouter()

	body, _ := json.Marshal(map[string]any{"menu_url": "https://example.com/menu"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/restaurants/nonexistent", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "restaurantId", Value: "nonexistent"}}

	handler.Update(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
