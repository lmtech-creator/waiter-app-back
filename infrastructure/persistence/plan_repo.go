package persistence

import (
	"github.com/waiter/back/domain/entity"
	"gorm.io/gorm"
)

type PlanRepo struct {
	db *gorm.DB
}

func NewPlanRepo(db *gorm.DB) *PlanRepo {
	return &PlanRepo{db: db}
}

func (r *PlanRepo) Create(plan *entity.Plan) error {
	return r.db.Create(plan).Error
}

func (r *PlanRepo) FindByID(id string) (*entity.Plan, error) {
	var plan entity.Plan
	if err := r.db.First(&plan, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *PlanRepo) FindByName(name string) (*entity.Plan, error) {
	var plan entity.Plan
	if err := r.db.First(&plan, "lower(name) = lower(?)", name).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *PlanRepo) FindAll() ([]entity.Plan, error) {
	var plans []entity.Plan
	if err := r.db.Order("name ASC").Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *PlanRepo) Update(plan *entity.Plan) error {
	return r.db.Save(plan).Error
}

func (r *PlanRepo) DeleteByID(id string) error {
	return r.db.Delete(&entity.Plan{}, "id = ?", id).Error
}

func (r *PlanRepo) CountRestaurants(planID string) (int64, error) {
	var count int64
	err := r.db.Model(&entity.Restaurant{}).Where("plan_id = ?", planID).Count(&count).Error
	return count, err
}
