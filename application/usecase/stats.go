package usecase

import (
	"fmt"
	"time"

	"github.com/waiter/back/domain/repository"
)

type StatsUseCase struct {
	requestRepo  repository.RequestRepository
	feedbackRepo repository.FeedbackRepository
}

func NewStatsUseCase(rr repository.RequestRepository, fr repository.FeedbackRepository) *StatsUseCase {
	return &StatsUseCase{requestRepo: rr, feedbackRepo: fr}
}

type RestaurantStats struct {
	TotalTablesServed  int                      `json:"total_tables_served"`
	AvgServiceTimeSec  float64                  `json:"avg_service_time_seconds"`
	UsageByHour       []repository.HourCount   `json:"usage_by_hour"`
	ScoreByHour       []repository.HourScore   `json:"score_by_hour"`
}

func (uc *StatsUseCase) GetStats(restaurantID string, since, until time.Time) (*RestaurantStats, error) {
	if since.IsZero() || until.IsZero() {
		return nil, fmt.Errorf("since and until are required")
	}
	if until.Before(since) {
		return nil, fmt.Errorf("until must be after since")
	}

	tablesServed, err := uc.requestRepo.CountServedTables(restaurantID, since, until)
	if err != nil {
		return nil, fmt.Errorf("counting served tables: %w", err)
	}

	avgTime, err := uc.requestRepo.AvgServiceTime(restaurantID, since, until)
	if err != nil {
		return nil, fmt.Errorf("computing avg service time: %w", err)
	}

	usage, err := uc.requestRepo.UsageByHour(restaurantID, since, until)
	if err != nil {
		return nil, fmt.Errorf("computing usage by hour: %w", err)
	}

	scores, err := uc.feedbackRepo.AvgScoreByHour(restaurantID, since, until)
	if err != nil {
		return nil, fmt.Errorf("computing score by hour: %w", err)
	}

	if usage == nil {
		usage = []repository.HourCount{}
	}
	if scores == nil {
		scores = []repository.HourScore{}
	}

	return &RestaurantStats{
		TotalTablesServed: tablesServed,
		AvgServiceTimeSec: avgTime,
		UsageByHour:       usage,
		ScoreByHour:       scores,
	}, nil
}
