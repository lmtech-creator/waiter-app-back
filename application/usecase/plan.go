package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/domain/repository"
)

var (
	ErrPlanNotFound       = errors.New("plan not found")
	ErrPlanNameRequired   = errors.New("plan name is required")
	ErrPlanNameExists     = errors.New("plan name already exists")
	ErrPlanInUse          = errors.New("plan is assigned to one or more restaurants")
)

type PlanUseCase struct {
	planRepo repository.PlanRepository
}

func NewPlanUseCase(pr repository.PlanRepository) *PlanUseCase {
	return &PlanUseCase{planRepo: pr}
}

type CreatePlanInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdatePlanInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (uc *PlanUseCase) CreatePlan(input CreatePlanInput) (*entity.Plan, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrPlanNameRequired
	}
	if _, err := uc.planRepo.FindByName(name); err == nil {
		return nil, ErrPlanNameExists
	}

	plan := &entity.Plan{
		ID:          uuid.New().String(),
		Name:        name,
		Description: strings.TrimSpace(input.Description),
	}
	if err := uc.planRepo.Create(plan); err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}
	return plan, nil
}

func (uc *PlanUseCase) GetPlan(id string) (*entity.Plan, error) {
	plan, err := uc.planRepo.FindByID(id)
	if err != nil {
		return nil, ErrPlanNotFound
	}
	return plan, nil
}

func (uc *PlanUseCase) GetAllPlans() ([]entity.Plan, error) {
	return uc.planRepo.FindAll()
}

func (uc *PlanUseCase) UpdatePlan(id string, input UpdatePlanInput) (*entity.Plan, error) {
	plan, err := uc.planRepo.FindByID(id)
	if err != nil {
		return nil, ErrPlanNotFound
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, ErrPlanNameRequired
		}
		existing, err := uc.planRepo.FindByName(name)
		if err == nil && existing.ID != plan.ID {
			return nil, ErrPlanNameExists
		}
		plan.Name = name
	}
	if input.Description != nil {
		plan.Description = strings.TrimSpace(*input.Description)
	}

	if err := uc.planRepo.Update(plan); err != nil {
		return nil, fmt.Errorf("failed to update plan: %w", err)
	}
	return plan, nil
}

func (uc *PlanUseCase) DeletePlan(id string) error {
	if _, err := uc.planRepo.FindByID(id); err != nil {
		return ErrPlanNotFound
	}
	count, err := uc.planRepo.CountRestaurants(id)
	if err != nil {
		return fmt.Errorf("failed to check plan usage: %w", err)
	}
	if count > 0 {
		return ErrPlanInUse
	}
	if err := uc.planRepo.DeleteByID(id); err != nil {
		return fmt.Errorf("failed to delete plan: %w", err)
	}
	return nil
}

// EnsureDefaultPlan returns the default "free" plan, creating it if missing.
func (uc *PlanUseCase) EnsureDefaultPlan() (*entity.Plan, error) {
	plan, err := uc.planRepo.FindByName("free")
	if err == nil {
		return plan, nil
	}
	plan = &entity.Plan{
		ID:          uuid.New().String(),
		Name:        "free",
		Description: "Plan gratuito",
	}
	if err := uc.planRepo.Create(plan); err != nil {
		// Handle race/duplicate: try to read again.
		existing, findErr := uc.planRepo.FindByName("free")
		if findErr == nil {
			return existing, nil
		}
		return nil, fmt.Errorf("failed to ensure default plan: %w", err)
	}
	return plan, nil
}
