package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/application/usecase"
	"github.com/waiter/back/domain/repository"
	"github.com/waiter/back/interfaces/http/dto"
	"github.com/waiter/back/mocks"
)

func setupStatsRouter() (*StatsHandler, *mocks.RequestRepo, *mocks.FeedbackRepo) {
	rr := mocks.NewRequestRepo()
	fr := mocks.NewFeedbackRepo()
	uc := usecase.NewStatsUseCase(rr, fr)
	handler := NewStatsHandler(uc)
	return handler, rr, fr
}

func TestStatsHandler_Get_Success(t *testing.T) {
	handler, rr, fr := setupStatsRouter()

	rr.CountServedTablesFn = func(rid string, s, u time.Time) (int, error) {
		return 42, nil
	}
	rr.AvgServiceTimeFn = func(rid string, s, u time.Time) (float64, error) {
		return 15.5, nil
	}
	rr.UsageByHourFn = func(rid string, s, u time.Time) ([]repository.HourCount, error) {
		return []repository.HourCount{{Hour: 14, Count: 10}}, nil
	}
	fr.AvgScoreByHourFn = func(rid string, s, u time.Time) ([]repository.HourScore, error) {
		return []repository.HourScore{{Hour: 14, AvgScore: 4.5}}, nil
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/r1/stats?since=2026-01-01T00:00:00Z&until=2026-01-02T00:00:00Z", nil)
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}}

	handler.Get(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp dto.StatsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.TotalTablesServed != 42 {
		t.Errorf("expected 42 tables, got %d", resp.TotalTablesServed)
	}
	if resp.AvgServiceTimeSec != 15.5 {
		t.Errorf("expected 15.5, got %f", resp.AvgServiceTimeSec)
	}
	if len(resp.UsageByHour) != 1 || resp.UsageByHour[0].Hour != 14 {
		t.Errorf("expected 1 usage entry for hour 14")
	}
	if len(resp.ScoreByHour) != 1 || resp.ScoreByHour[0].Hour != 14 {
		t.Errorf("expected 1 score entry for hour 14")
	}
}

func TestStatsHandler_Get_MissingSince(t *testing.T) {
	handler, _, _ := setupStatsRouter()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/r1/stats", nil)
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}}

	handler.Get(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestStatsHandler_Get_MissingUntil(t *testing.T) {
	handler, _, _ := setupStatsRouter()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/r1/stats?since=2026-01-01T00:00:00Z", nil)
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}}

	handler.Get(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestStatsHandler_Get_BadSince(t *testing.T) {
	handler, _, _ := setupStatsRouter()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/r1/stats?since=invalid&until=2026-01-02T00:00:00Z", nil)
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}}

	handler.Get(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestStatsHandler_Get_UseCaseError(t *testing.T) {
	handler, rr, _ := setupStatsRouter()

	rr.CountServedTablesFn = func(rid string, s, u time.Time) (int, error) {
		return 0, fmt.Errorf("db error")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/restaurants/r1/stats?since=2026-01-01T00:00:00Z&until=2026-01-02T00:00:00Z", nil)
	c.Params = gin.Params{{Key: "restaurantId", Value: "r1"}}

	handler.Get(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
