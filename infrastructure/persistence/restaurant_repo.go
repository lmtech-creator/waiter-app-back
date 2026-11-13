package persistence

import (
	"github.com/waiter/back/domain/entity"
	"gorm.io/gorm"
)

type RestaurantRepo struct {
	db *gorm.DB
}

func NewRestaurantRepo(db *gorm.DB) *RestaurantRepo {
	return &RestaurantRepo{db: db}
}

func (r *RestaurantRepo) Create(restaurant *entity.Restaurant) error {
	return r.db.Create(restaurant).Error
}

func (r *RestaurantRepo) FindByID(id string) (*entity.Restaurant, error) {
	var restaurant entity.Restaurant
	if err := r.db.Preload("Plan").First(&restaurant, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &restaurant, nil
}

func (r *RestaurantRepo) FindAll() ([]entity.Restaurant, error) {
	var restaurants []entity.Restaurant
	if err := r.db.Preload("Plan").Find(&restaurants).Error; err != nil {
		return nil, err
	}
	return restaurants, nil
}
