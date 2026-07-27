package persistence

import (
	"github.com/waiter/back/domain/entity"
	"gorm.io/gorm"
)

type AdminRepo struct {
	db *gorm.DB
}

func NewAdminRepo(db *gorm.DB) *AdminRepo {
	return &AdminRepo{db: db}
}

func (r *AdminRepo) FindByUsername(username string) (*entity.AdminUser, error) {
	var admin entity.AdminUser
	if err := r.db.First(&admin, "username = ?", username).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepo) FindByID(id string) (*entity.AdminUser, error) {
	var admin entity.AdminUser
	if err := r.db.First(&admin, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepo) FindByRestaurantID(restaurantID string) ([]entity.AdminUser, error) {
	var admins []entity.AdminUser
	if err := r.db.Where("restaurant_id = ?", restaurantID).Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}

func (r *AdminRepo) FindAll() ([]entity.AdminUser, error) {
	var admins []entity.AdminUser
	if err := r.db.Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}

func (r *AdminRepo) Create(admin *entity.AdminUser) error {
	return r.db.Create(admin).Error
}

func (r *AdminRepo) ExistsAny() (bool, error) {
	var count int64
	if err := r.db.Model(&entity.AdminUser{}).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *AdminRepo) Update(admin *entity.AdminUser) error {
	return r.db.Save(admin).Error
}

func (r *AdminRepo) DeleteByID(id string) error {
	return r.db.Delete(&entity.AdminUser{}, "id = ?", id).Error
}
