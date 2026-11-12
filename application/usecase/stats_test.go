package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/waiter/back/domain/repository"
	"github.com/waiter/back/mocks"
)

func TestStatsUseCase_GetStats_Valid(t *testing.T) {
	rr := mocks.NewRequestRepo()
	fr := mocks.NewFeedbackRepo()

	rr.CountServedTablesFn = func(rid string, s, u time.Time) (int, error) {
		return 42, nil
	}
	rr.AvgServiceTimeFn = func(rid string, s, u time.Time) (float64, error) {
		return 15.5, nil
	}
	rr.UsageByHourFn = func(rid string, s, u time.Time) ([]repository.HourBucket, error) {
		// 17:00 UTC is 14:00 in Buenos Aires (UTC-3).
		return []repository.HourBucket{{Start: time.Date(2026, 1, 1, 17, 0, 0, 0, time.UTC), Count: 10}}, nil
	}
	fr.AvgScoreByHourFn = func(rid string, s, u time.Time) ([]repository.ScoreBucket, error) {
		return []repository.ScoreBucket{{Start: time.Date(2026, 1, 1, 17, 0, 0, 0, time.UTC), ScoreSum: 45, Count: 10}}, nil
	}

	uc := NewStatsUseCase(rr, fr)
	since, _ := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
	until, _ := time.Parse(time.RFC3339, "2026-01-02T00:00:00Z")

	stats, err := uc.GetStats("rest-1", since, until)
	require.NoError(t, err)
	require.Equal(t, 42, stats.TotalTablesServed)
	require.Equal(t, 15.5, stats.AvgServiceTimeSec)
	require.Len(t, stats.UsageByHour, 1)
	require.Equal(t, 14, stats.UsageByHour[0].Hour)
	require.Equal(t, 10, stats.UsageByHour[0].Count)
	require.Len(t, stats.ScoreByHour, 1)
	require.Equal(t, 14, stats.ScoreByHour[0].Hour)
	require.Equal(t, 4.5, stats.ScoreByHour[0].AvgScore)
}

func TestStatsUseCase_GetStats_NoData(t *testing.T) {
	rr := mocks.NewRequestRepo()
	fr := mocks.NewFeedbackRepo()
	uc := NewStatsUseCase(rr, fr)

	since, _ := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
	until, _ := time.Parse(time.RFC3339, "2026-01-02T00:00:00Z")

	stats, err := uc.GetStats("rest-1", since, until)
	require.NoError(t, err)
	require.Equal(t, 0, stats.TotalTablesServed)
	require.Equal(t, float64(0), stats.AvgServiceTimeSec)
	require.Empty(t, stats.UsageByHour)
	require.Empty(t, stats.ScoreByHour)
}

func TestStatsUseCase_GetStats_Validation(t *testing.T) {
	rr := mocks.NewRequestRepo()
	fr := mocks.NewFeedbackRepo()
	uc := NewStatsUseCase(rr, fr)

	t.Run("since zero", func(t *testing.T) {
		_, err := uc.GetStats("rest-1", time.Time{}, time.Now())
		require.Error(t, err)
	})

	t.Run("until zero", func(t *testing.T) {
		_, err := uc.GetStats("rest-1", time.Now(), time.Time{})
		require.Error(t, err)
	})

	t.Run("until before since", func(t *testing.T) {
		since, _ := time.Parse(time.RFC3339, "2026-01-02T00:00:00Z")
		until, _ := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
		_, err := uc.GetStats("rest-1", since, until)
		require.Error(t, err)
	})
}

func TestStatsUseCase_GetStats_RepoError(t *testing.T) {
	rr := mocks.NewRequestRepo()
	fr := mocks.NewFeedbackRepo()

	rr.CountServedTablesFn = func(rid string, s, u time.Time) (int, error) {
		return 0, assert.AnError
	}

	uc := NewStatsUseCase(rr, fr)
	since, _ := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
	until, _ := time.Parse(time.RFC3339, "2026-01-02T00:00:00Z")

	_, err := uc.GetStats("rest-1", since, until)
	require.Error(t, err)
}

// Regression: a 20:00 Buenos Aires request is 23:00 UTC, and it must be
// reported as hour 20, never 23 (UTC) nor 17 (double shift).
func TestUsageByLocalHour_UsesBuenosAiresOffset(t *testing.T) {
	buckets := []repository.HourBucket{
		{Start: time.Date(2026, 11, 12, 23, 0, 0, 0, time.UTC), Count: 3},
		{Start: time.Date(2026, 11, 13, 23, 0, 0, 0, time.UTC), Count: 2},
		{Start: time.Date(2026, 11, 12, 17, 0, 0, 0, time.UTC), Count: 1},
	}

	require.Equal(t, []repository.HourCount{
		{Hour: 14, Count: 1},
		{Hour: 20, Count: 5},
	}, usageByLocalHour(buckets))
}

func TestScoresByLocalHour_MergesDaysWithWeightedAverage(t *testing.T) {
	buckets := []repository.ScoreBucket{
		{Start: time.Date(2026, 11, 12, 23, 0, 0, 0, time.UTC), ScoreSum: 10, Count: 2},
		{Start: time.Date(2026, 11, 13, 23, 0, 0, 0, time.UTC), ScoreSum: 12, Count: 3},
	}

	require.Equal(t, []repository.HourScore{
		{Hour: 20, AvgScore: 4.4},
	}, scoresByLocalHour(buckets))
}

func TestUsageByLocalHour_Empty(t *testing.T) {
	require.Empty(t, usageByLocalHour(nil))
	require.Empty(t, scoresByLocalHour(nil))
}
