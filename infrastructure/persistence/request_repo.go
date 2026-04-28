package persistence

import (
	"github.com/waiter/back/domain/entity"
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
