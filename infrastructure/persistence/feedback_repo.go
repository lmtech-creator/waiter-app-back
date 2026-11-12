package persistence

import (
	"time"

	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/domain/repository"
	"gorm.io/gorm"
)

type FeedbackRepo struct {
	db *gorm.DB
}

func NewFeedbackRepo(db *gorm.DB) *FeedbackRepo {
	return &FeedbackRepo{db: db}
}

func (r *FeedbackRepo) Create(feedback *entity.Feedback) error {
	return r.db.Create(feedback).Error
}

func (r *FeedbackRepo) AvgScoreByHour(restaurantID string, since, until time.Time) ([]repository.ScoreBucket, error) {
	var results []repository.ScoreBucket
	// Same UTC-epoch bucketing as UsageByHour; SUM + COUNT let the use case
	// compute a weighted average after merging into local hours.
	err := r.db.Model(&entity.Feedback{}).
		Joins("JOIN tables ON tables.id = feedbacks.table_id").
		Where("tables.restaurant_id = ? AND feedbacks.created_at BETWEEN ? AND ?", restaurantID, since, until).
		Select("to_timestamp(floor((EXTRACT(EPOCH FROM feedbacks.created_at))::double precision / 3600) * 3600) AS start, SUM(feedbacks.score)::double precision AS score_sum, COUNT(*) AS count").
		Group("start").
		Order("start").
		Scan(&results).Error
	return results, err
}

func (r *FeedbackRepo) FindByRestaurantID(restaurantID string) ([]entity.Feedback, error) {
	var feedbacks []entity.Feedback
	if err := r.db.Joins("JOIN tables ON tables.id = feedbacks.table_id").
		Where("tables.restaurant_id = ?", restaurantID).
		Order("feedbacks.created_at DESC").
		Find(&feedbacks).Error; err != nil {
		return nil, err
	}
	return feedbacks, nil
}

func (r *FeedbackRepo) FindByTableID(tableID string) ([]entity.Feedback, error) {
	var feedbacks []entity.Feedback
	if err := r.db.Where("table_id = ?", tableID).Order("created_at DESC").Find(&feedbacks).Error; err != nil {
		return nil, err
	}
	return feedbacks, nil
}
