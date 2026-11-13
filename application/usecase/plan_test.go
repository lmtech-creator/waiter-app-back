package usecase

import (
	"errors"
	"fmt"
	"testing"

	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/mocks"
)

func newPlanUseCase() (*PlanUseCase, *mocks.PlanRepo) {
	planRepo := mocks.NewPlanRepo()
	return NewPlanUseCase(planRepo), planRepo
}

func TestCreatePlan_Success(t *testing.T) {
	uc, planRepo := newPlanUseCase()

	plan, err := uc.CreatePlan(CreatePlanInput{Name: "  premium  ", Description: "  Plan premium  "})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if plan.Name != "premium" {
		t.Errorf("expected trimmed name 'premium', got %q", plan.Name)
	}
	if plan.Description != "Plan premium" {
		t.Errorf("expected trimmed description, got %q", plan.Description)
	}
	if plan.ID == "" {
		t.Error("expected non-empty ID")
	}
	if _, ok := planRepo.Plans[plan.ID]; !ok {
		t.Error("expected plan to be persisted in repo")
	}
}

func TestCreatePlan_NameRequired(t *testing.T) {
	uc, _ := newPlanUseCase()

	_, err := uc.CreatePlan(CreatePlanInput{Name: "   "})
	if !errors.Is(err, ErrPlanNameRequired) {
		t.Fatalf("expected ErrPlanNameRequired, got %v", err)
	}
}

func TestCreatePlan_NameExists(t *testing.T) {
	uc, planRepo := newPlanUseCase()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "pro"}

	_, err := uc.CreatePlan(CreatePlanInput{Name: "PRO"})
	if !errors.Is(err, ErrPlanNameExists) {
		t.Fatalf("expected ErrPlanNameExists, got %v", err)
	}
}

func TestGetPlan_Success(t *testing.T) {
	uc, planRepo := newPlanUseCase()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "pro"}

	plan, err := uc.GetPlan("p1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if plan.Name != "pro" {
		t.Errorf("expected plan 'pro', got %q", plan.Name)
	}
}

func TestGetPlan_NotFound(t *testing.T) {
	uc, _ := newPlanUseCase()

	_, err := uc.GetPlan("missing")
	if !errors.Is(err, ErrPlanNotFound) {
		t.Fatalf("expected ErrPlanNotFound, got %v", err)
	}
}

func TestGetAllPlans(t *testing.T) {
	uc, planRepo := newPlanUseCase()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "free"}
	planRepo.Plans["p2"] = &entity.Plan{ID: "p2", Name: "pro"}

	plans, err := uc.GetAllPlans()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(plans) != 2 {
		t.Errorf("expected 2 plans, got %d", len(plans))
	}
}

func TestUpdatePlan_Success(t *testing.T) {
	uc, planRepo := newPlanUseCase()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "pro", Description: "old"}

	name := "premium"
	description := "new"
	plan, err := uc.UpdatePlan("p1", UpdatePlanInput{Name: &name, Description: &description})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if plan.Name != "premium" || plan.Description != "new" {
		t.Errorf("expected updated plan, got name=%q description=%q", plan.Name, plan.Description)
	}
}

func TestUpdatePlan_NotFound(t *testing.T) {
	uc, _ := newPlanUseCase()

	name := "premium"
	_, err := uc.UpdatePlan("missing", UpdatePlanInput{Name: &name})
	if !errors.Is(err, ErrPlanNotFound) {
		t.Fatalf("expected ErrPlanNotFound, got %v", err)
	}
}

func TestUpdatePlan_NameRequired(t *testing.T) {
	uc, planRepo := newPlanUseCase()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "pro"}

	name := "   "
	_, err := uc.UpdatePlan("p1", UpdatePlanInput{Name: &name})
	if !errors.Is(err, ErrPlanNameRequired) {
		t.Fatalf("expected ErrPlanNameRequired, got %v", err)
	}
}

func TestUpdatePlan_NameExists(t *testing.T) {
	uc, planRepo := newPlanUseCase()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "pro"}
	planRepo.Plans["p2"] = &entity.Plan{ID: "p2", Name: "premium"}

	name := "premium"
	_, err := uc.UpdatePlan("p1", UpdatePlanInput{Name: &name})
	if !errors.Is(err, ErrPlanNameExists) {
		t.Fatalf("expected ErrPlanNameExists, got %v", err)
	}
}

func TestUpdatePlan_SameNameAllowed(t *testing.T) {
	uc, planRepo := newPlanUseCase()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "pro"}

	name := "pro"
	if _, err := uc.UpdatePlan("p1", UpdatePlanInput{Name: &name}); err != nil {
		t.Fatalf("expected no error updating a plan with its own name, got %v", err)
	}
}

func TestDeletePlan_Success(t *testing.T) {
	uc, planRepo := newPlanUseCase()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "pro"}

	if err := uc.DeletePlan("p1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, ok := planRepo.Plans["p1"]; ok {
		t.Error("expected plan to be deleted")
	}
}

func TestDeletePlan_NotFound(t *testing.T) {
	uc, _ := newPlanUseCase()

	err := uc.DeletePlan("missing")
	if !errors.Is(err, ErrPlanNotFound) {
		t.Fatalf("expected ErrPlanNotFound, got %v", err)
	}
}

func TestDeletePlan_InUse(t *testing.T) {
	uc, planRepo := newPlanUseCase()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "pro"}
	planRepo.CountRestaurantsFn = func(planID string) (int64, error) {
		return 2, nil
	}

	err := uc.DeletePlan("p1")
	if !errors.Is(err, ErrPlanInUse) {
		t.Fatalf("expected ErrPlanInUse, got %v", err)
	}
}

func TestDeletePlan_CountError(t *testing.T) {
	uc, planRepo := newPlanUseCase()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "pro"}
	planRepo.CountRestaurantsFn = func(planID string) (int64, error) {
		return 0, fmt.Errorf("db error")
	}

	if err := uc.DeletePlan("p1"); err == nil {
		t.Fatal("expected error from repo failure")
	}
}

func TestEnsureDefaultPlan_CreatesWhenMissing(t *testing.T) {
	uc, planRepo := newPlanUseCase()

	plan, err := uc.EnsureDefaultPlan()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if plan.Name != "free" {
		t.Errorf("expected default plan 'free', got %q", plan.Name)
	}
	if len(planRepo.Plans) != 1 {
		t.Errorf("expected 1 plan in repo, got %d", len(planRepo.Plans))
	}
}

func TestEnsureDefaultPlan_ReturnsExisting(t *testing.T) {
	uc, planRepo := newPlanUseCase()
	planRepo.Plans["p1"] = &entity.Plan{ID: "p1", Name: "free"}

	plan, err := uc.EnsureDefaultPlan()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if plan.ID != "p1" {
		t.Errorf("expected existing plan p1, got %q", plan.ID)
	}
	if len(planRepo.Plans) != 1 {
		t.Errorf("expected no new plan to be created, got %d", len(planRepo.Plans))
	}
}
