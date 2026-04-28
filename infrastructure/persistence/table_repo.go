package persistence

import (
	"github.com/waiter/back/domain/entity"
	"gorm.io/gorm"
)

type TableRepo struct {
	db *gorm.DB
}

func NewTableRepo(db *gorm.DB) *TableRepo {
	return &TableRepo{db: db}
}

func (r *TableRepo) Create(table *entity.Table) error {
	return r.db.Create(table).Error
}

func (r *TableRepo) FindByID(id string) (*entity.Table, error) {
	var table entity.Table
	if err := r.db.First(&table, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &table, nil
}

func (r *TableRepo) FindByRestaurantID(restaurantID string) ([]entity.Table, error) {
	var tables []entity.Table
	if err := r.db.Where("restaurant_id = ?", restaurantID).Find(&tables).Error; err != nil {
		return nil, err
	}
	return tables, nil
}

func (r *TableRepo) FindByQRCode(qrCode string) (*entity.Table, error) {
	var table entity.Table
	if err := r.db.First(&table, "qr_code = ?", qrCode).Error; err != nil {
		return nil, err
	}
	return &table, nil
}

func (r *TableRepo) UpdateQRCode(id, qrCode string) error {
	return r.db.Model(&entity.Table{}).Where("id = ?", id).Update("qr_code", qrCode).Error
}
