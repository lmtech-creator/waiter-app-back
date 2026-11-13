package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/application/usecase"
	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/interfaces/http/dto"
	"github.com/waiter/back/mocks"
)

func setupPlanHandler() (*PlanHandler, *mocks.PlanRepo) {
	planRepo := mocks.NewPlanRepo()
	uc := usecase.NewPlanUseCase(planRepo)
	return NewPlanHandler(uc), planRepo
}

func planRequest(method, path string, body []byte, params gin.Params) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var reader *bytes.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	} else {
		reader = bytes.NewReader(nil)
	}
	c.Request = httptest.NewRequest(method, path, reader)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = params
	return w, c
}

func TestPlanHandler_Create_Success(t *testing.T) {
	handler, _ := setupPlanHandler()

	body, _ := json.Marshal(map[string]any{"name": "pro", "description": "Plan pro"})
	w, c := planRequest(http.MethodPost, "/api/v1/plans", body, nil)

	handler.Create(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var resp dto.PlanResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Name != "pro" || resp.Description != "Plan pro" {
		t.Errorf("unexpected plan response: %+v", resp)
	}
	if resp.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestPlanHandler_Create_BadJSON(t *testing.T) {
	handler, _ := setupPlanHandler()

	w, c := planRequest(http.MethodPost, "/api/v1/plans", []byte("{"), nil)

	handler.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPlanHandler_Create_NameRequired(t *testing.T) {
	handler, _ := setupPlanHandler()

	body, _ := json.Marshal(map[string]any{"name": "   "})
	w, c := planRequest(http.MethodPost, "/api/v1/plans", body, nil)

	handler.Create(c)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}

func TestPlanHandler_Create_NameExists(t *testing.T) {
	handler, planRepo := setupPlanHandler()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "pro"}

	body, _ := json.Marshal(map[string]any{"name": "pro"})
	w, c := planRequest(http.MethodPost, "/api/v1/plans", body, nil)

	handler.Create(c)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestPlanHandler_GetAll_Success(t *testing.T) {
	handler, planRepo := setupPlanHandler()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "free"}
	planRepo.Plans["p2"] = &entity.Plan{ID: "p2", Name: "pro"}

	w, c := planRequest(http.MethodGet, "/api/v1/plans", nil, nil)

	handler.GetAll(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []dto.PlanResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 plans, got %d", len(resp))
	}
}

func TestPlanHandler_Get_Success(t *testing.T) {
	handler, planRepo := setupPlanHandler()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "pro", Description: "Plan pro"}

	w, c := planRequest(http.MethodGet, "/api/v1/plans/p1", nil, gin.Params{{Key: "planId", Value: "p1"}})

	handler.Get(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp dto.PlanResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Name != "pro" {
		t.Errorf("expected plan 'pro', got %q", resp.Name)
	}
}

func TestPlanHandler_Get_NotFound(t *testing.T) {
	handler, _ := setupPlanHandler()

	w, c := planRequest(http.MethodGet, "/api/v1/plans/missing", nil, gin.Params{{Key: "planId", Value: "missing"}})

	handler.Get(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPlanHandler_Update_Success(t *testing.T) {
	handler, planRepo := setupPlanHandler()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "pro"}

	body, _ := json.Marshal(map[string]any{"name": "premium"})
	w, c := planRequest(http.MethodPatch, "/api/v1/plans/p1", body, gin.Params{{Key: "planId", Value: "p1"}})

	handler.Update(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp dto.PlanResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Name != "premium" {
		t.Errorf("expected plan 'premium', got %q", resp.Name)
	}
}

func TestPlanHandler_Update_NotFound(t *testing.T) {
	handler, _ := setupPlanHandler()

	body, _ := json.Marshal(map[string]any{"name": "premium"})
	w, c := planRequest(http.MethodPatch, "/api/v1/plans/missing", body, gin.Params{{Key: "planId", Value: "missing"}})

	handler.Update(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPlanHandler_Update_NameRequired(t *testing.T) {
	handler, planRepo := setupPlanHandler()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "pro"}

	body, _ := json.Marshal(map[string]any{"name": "   "})
	w, c := planRequest(http.MethodPatch, "/api/v1/plans/p1", body, gin.Params{{Key: "planId", Value: "p1"}})

	handler.Update(c)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}

func TestPlanHandler_Update_NameExists(t *testing.T) {
	handler, planRepo := setupPlanHandler()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "pro"}
	planRepo.Plans["p2"] = &entity.Plan{ID: "p2", Name: "premium"}

	body, _ := json.Marshal(map[string]any{"name": "premium"})
	w, c := planRequest(http.MethodPatch, "/api/v1/plans/p1", body, gin.Params{{Key: "planId", Value: "p1"}})

	handler.Update(c)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestPlanHandler_Delete_Success(t *testing.T) {
	handler, planRepo := setupPlanHandler()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "pro"}

	// Use a real router so gin flushes the 204 status (a bare test context
	// defers the write until the engine's WriteHeaderNow).
	w := httptest.NewRecorder()
	r := gin.New()
	r.DELETE("/api/v1/plans/:planId", handler.Delete)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/plans/p1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if _, ok := planRepo.Plans["p1"]; ok {
		t.Error("expected plan to be deleted")
	}
}

func TestPlanHandler_Delete_NotFound(t *testing.T) {
	handler, _ := setupPlanHandler()

	w, c := planRequest(http.MethodDelete, "/api/v1/plans/missing", nil, gin.Params{{Key: "planId", Value: "missing"}})

	handler.Delete(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPlanHandler_Delete_InUse(t *testing.T) {
	handler, planRepo := setupPlanHandler()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "pro"}
	planRepo.CountRestaurantsFn = func(planID string) (int64, error) {
		return 1, nil
	}

	w, c := planRequest(http.MethodDelete, "/api/v1/plans/p1", nil, gin.Params{{Key: "planId", Value: "p1"}})

	handler.Delete(c)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}
