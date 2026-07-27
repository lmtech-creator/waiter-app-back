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
	rr.UsageByHourFn = func(rid string, s, u time.Time) ([]repository.HourCount, error) {
		return []repository.HourCount{{Hour: 14, Count: 10}}, nil
	}
	fr.AvgScoreByHourFn = func(rid string, s, u time.Time) ([]repository.HourScore, error) {
		return []repository.HourScore{{Hour: 14, AvgScore: 4.5}}, nil
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
