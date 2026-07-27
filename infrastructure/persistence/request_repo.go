package persistence

import (
	"time"

	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/domain/repository"
	"gorm.io/gorm"
)

type RequestRepo struct {
	db *gorm.DB
}

func NewRequestRepo(db *gorm.DB) *RequestRepo {
	return &RequestRepo{db: db}
}

func (r *RequestRepo) Create(request *entity.Request) error {
	return r.db.Create(request).Error
}

func (r *RequestRepo) FindByID(id string) (*entity.Request, error) {
	var request entity.Request
	if err := r.db.First(&request, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *RequestRepo) FindActiveByRestaurantID(restaurantID string) ([]entity.Request, error) {
	var requests []entity.Request
	err := r.db.
		Joins("JOIN tables ON tables.id = requests.table_id").
		Where("tables.restaurant_id = ? AND requests.status != ?", restaurantID, entity.Done).
		Find(&requests).Error
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *RequestRepo) FindByTableID(tableID string) ([]entity.Request, error) {
	var requests []entity.Request
	if err := r.db.Where("table_id = ? AND status != ?", tableID, entity.Done).Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *RequestRepo) UpdateStatus(id string, status entity.RequestStatus) error {
	return r.db.Model(&entity.Request{}).Where("id = ?", id).Update("status", status).Error
}

func (r *RequestRepo) UpdateCompletedAt(id string, t time.Time) error {
	return r.db.Model(&entity.Request{}).Where("id = ?", id).Update("completed_at", t).Error
}

func (r *RequestRepo) CountServedTables(restaurantID string, since, until time.Time) (int, error) {
	var count int64
	err := r.db.Model(&entity.Request{}).
		Joins("JOIN tables ON tables.id = requests.table_id").
		Where("tables.restaurant_id = ? AND requests.status = ? AND requests.created_at BETWEEN ? AND ?", restaurantID, entity.Done, since, until).
		Distinct("requests.table_id").
		Count(&count).Error
	return int(count), err
}

func (r *RequestRepo) AvgServiceTime(restaurantID string, since, until time.Time) (float64, error) {
	var avg float64
	err := r.db.Model(&entity.Request{}).
		Joins("JOIN tables ON tables.id = requests.table_id").
		Where("tables.restaurant_id = ? AND requests.status = ? AND requests.completed_at IS NOT NULL AND requests.created_at BETWEEN ? AND ?", restaurantID, entity.Done, since, until).
		Select("COALESCE(AVG(EXTRACT(EPOCH FROM (requests.completed_at - requests.created_at))), 0)").
		Scan(&avg).Error
	return avg, err
}

func (r *RequestRepo) UsageByHour(restaurantID string, since, until time.Time) ([]repository.HourCount, error) {
	var results []repository.HourCount
	err := r.db.Model(&entity.Request{}).
		Joins("JOIN tables ON tables.id = requests.table_id").
		Where("tables.restaurant_id = ? AND requests.created_at BETWEEN ? AND ?", restaurantID, since, until).
		Select("EXTRACT(HOUR FROM requests.created_at)::int AS hour, COUNT(*) AS count").
		Group("hour").
		Order("hour").
		Scan(&results).Error
	return results, err
}

func (r *RequestRepo) FindLastCreatedByTableID(tableID string) (*entity.Request, error) {
	var request entity.Request
	err := r.db.Where("table_id = ?", tableID).
		Order("created_at DESC").
		First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}
