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

func init() {
	gin.SetMode(gin.TestMode)
}

func setupFeedbackRouter() (*FeedbackHandler, *mocks.FeedbackRepo, *mocks.TableRepo) {
	feedbackRepo := mocks.NewFeedbackRepo()
	tableRepo := mocks.NewTableRepo()
	uc := usecase.NewFeedbackUseCase(feedbackRepo, tableRepo)
	handler := NewFeedbackHandler(uc)
	return handler, feedbackRepo, tableRepo
}

func TestFeedbackHandler_Create_Success(t *testing.T) {
	handler, _, tableRepo := setupFeedbackRouter()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", RestaurantID: "r1"}

	body, _ := json.Marshal(map[string]any{"table_id": "t1", "score": 4, "comment": "Nice!"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/feedback", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	var resp dto.FeedbackResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Score != 4 {
		t.Errorf("expected score 4, got %d", resp.Score)
	}
}

func TestFeedbackHandler_Create_BadJSON(t *testing.T) {
	handler, _, _ := setupFeedbackRouter()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/feedback", bytes.NewReader([]byte(`{invalid`)))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestFeedbackHandler_Create_MissingFields(t *testing.T) {
	handler, _, _ := setupFeedbackRouter()

	body, _ := json.Marshal(map[string]any{"comment": "missing required fields"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/feedback", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestFeedbackHandler_Create_TableNotFound(t *testing.T) {
	handler, _, _ := setupFeedbackRouter()

	body, _ := json.Marshal(map[string]any{"table_id": "nonexistent", "score": 3})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/feedback", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
}

func TestFeedbackHandler_GetByRestaurant_Success(t *testing.T) {
	handler, feedbackRepo, _ := setupFeedbackRouter()
	feedbackRepo.Feedbacks["f1"] = &entity.Feedback{ID: "f1", TableID: "t1", Score: 5}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/r1/feedback", nil)
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}}

	handler.GetByRestaurant(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp []dto.FeedbackResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("expected 1 feedback, got %d", len(resp))
	}
}

func TestFeedbackHandler_GetByRestaurant_RepoError(t *testing.T) {
	handler, feedbackRepo, _ := setupFeedbackRouter()
	feedbackRepo.FindByRIDFn = func(restaurantID string) ([]entity.Feedback, error) {
		return nil, fmt.Errorf("db error")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/r1/feedback", nil)
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}}

	handler.GetByRestaurant(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
