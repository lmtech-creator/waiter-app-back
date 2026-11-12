package usecase

import (
	"fmt"
	"sort"
	"time"

	"github.com/waiter/back/domain/repository"
)

// reportLocation is the restaurant's reporting zone. Buenos Aires has been a
// fixed UTC-3 without DST since 2009, so a FixedZone keeps the bucketing
// deterministic: it does not depend on the host tzdata or on the Postgres
// session timezone.
var reportLocation = time.FixedZone("ART", -3*60*60)

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

	usageBuckets, err := uc.requestRepo.UsageByHour(restaurantID, since, until)
	if err != nil {
		return nil, fmt.Errorf("computing usage by hour: %w", err)
	}

	scoreBuckets, err := uc.feedbackRepo.AvgScoreByHour(restaurantID, since, until)
	if err != nil {
		return nil, fmt.Errorf("computing score by hour: %w", err)
	}

	return &RestaurantStats{
		TotalTablesServed: tablesServed,
		AvgServiceTimeSec: avgTime,
		UsageByHour:       usageByLocalHour(usageBuckets),
		ScoreByHour:       scoresByLocalHour(scoreBuckets),
	}, nil
}

// usageByLocalHour converts raw UTC hour buckets to the restaurant's local
// hour-of-day, merging buckets that land on the same hour (e.g. the same hour
// on different days). Result is ordered by hour.
func usageByLocalHour(buckets []repository.HourBucket) []repository.HourCount {
	counts := make(map[int]int, 24)
	for _, bucket := range buckets {
		counts[bucket.Start.In(reportLocation).Hour()] += bucket.Count
	}

	result := make([]repository.HourCount, 0, len(counts))
	for hour, count := range counts {
		result = append(result, repository.HourCount{Hour: hour, Count: count})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Hour < result[j].Hour })
	return result
}

// scoresByLocalHour converts raw UTC score buckets to local hours and computes
// a weighted average per hour, so merging days does not skew the score.
func scoresByLocalHour(buckets []repository.ScoreBucket) []repository.HourScore {
	sums := make(map[int]float64, 24)
	counts := make(map[int]int, 24)
	for _, bucket := range buckets {
		hour := bucket.Start.In(reportLocation).Hour()
		sums[hour] += bucket.ScoreSum
		counts[hour] += bucket.Count
	}

	result := make([]repository.HourScore, 0, len(counts))
	for hour, count := range counts {
		if count == 0 {
			continue
		}
		result = append(result, repository.HourScore{
			Hour:     hour,
			AvgScore: sums[hour] / float64(count),
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Hour < result[j].Hour })
	return result
}
