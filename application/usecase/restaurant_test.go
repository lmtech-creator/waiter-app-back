package usecase

import (
	"fmt"
	"testing"

	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/mocks"
)

func TestCreateRestaurant_Success(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	r, err := uc.CreateRestaurant(CreateRestaurantInput{Name: "Burger Place"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if r.Name != "Burger Place" {
		t.Errorf("expected name 'Burger Place', got %s", r.Name)
	}
	if r.Plan != "free" {
		t.Errorf("expected default plan 'free', got %s", r.Plan)
	}
	if r.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestCreateRestaurant_WithPlan(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	r, err := uc.CreateRestaurant(CreateRestaurantInput{Name: "Pizza House", Plan: "premium"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if r.Plan != "premium" {
		t.Errorf("expected plan 'premium', got %s", r.Plan)
	}
}

func TestCreateRestaurant_RepoError(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	restaurantRepo.CreateFn = func(r *entity.Restaurant) error {
		return fmt.Errorf("db error")
	}
	tableRepo := mocks.NewTableRepo()
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	_, err := uc.CreateRestaurant(CreateRestaurantInput{Name: "Test"})
	if err == nil {
		t.Fatal("expected error from repo failure")
	}
}

func TestGetRestaurant_Success(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	restaurantRepo.Restaurants["r1"] = &entity.Restaurant{ID: "r1", Name: "Sushi Bar", Plan: "free"}
	tableRepo := mocks.NewTableRepo()
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	r, err := uc.GetRestaurant("r1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if r.Name != "Sushi Bar" {
		t.Errorf("expected 'Sushi Bar', got %s", r.Name)
	}
}

func TestGetRestaurant_NotFound(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	_, err := uc.GetRestaurant("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent restaurant")
	}
}

func TestCreateTable_Success(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	restaurantRepo.Restaurants["r1"] = &entity.Restaurant{ID: "r1", Name: "Test"}
	tableRepo := mocks.NewTableRepo()
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	tbl, err := uc.CreateTable(CreateTableInput{Number: 5, RestaurantID: "r1"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tbl.Number != 5 {
		t.Errorf("expected number 5, got %d", tbl.Number)
	}
	if tbl.RestaurantID != "r1" {
		t.Errorf("expected restaurant_id r1, got %s", tbl.RestaurantID)
	}
	if tbl.QRCode == "" {
		t.Error("expected non-empty QR code")
	}
	if len(tbl.QRCode) != 10 {
		t.Errorf("expected QR code of length 10, got %d", len(tbl.QRCode))
	}
	if tbl.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestCreateTable_RestaurantNotFound(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	_, err := uc.CreateTable(CreateTableInput{Number: 1, RestaurantID: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent restaurant")
	}
}

func TestCreateTable_RepoError(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	restaurantRepo.Restaurants["r1"] = &entity.Restaurant{ID: "r1"}
	tableRepo := mocks.NewTableRepo()
	tableRepo.CreateFn = func(t *entity.Table) error {
		return fmt.Errorf("db error")
	}
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	_, err := uc.CreateTable(CreateTableInput{Number: 1, RestaurantID: "r1"})
	if err == nil {
		t.Fatal("expected error from repo failure")
	}
}

func TestGetTables_Success(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", RestaurantID: "r1", Number: 1}
	tableRepo.Tables["t2"] = &entity.Table{ID: "t2", RestaurantID: "r1", Number: 2}
	tableRepo.Tables["t3"] = &entity.Table{ID: "t3", RestaurantID: "r2", Number: 1}
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	tables, err := uc.GetTables("r1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tables) != 2 {
		t.Errorf("expected 2 tables, got %d", len(tables))
	}
}

func TestGetTables_Empty(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	tables, err := uc.GetTables("r99")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tables) != 0 {
		t.Errorf("expected 0 tables, got %d", len(tables))
	}
}

func TestGetTables_RepoError(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	tableRepo.FindByRestaurantFn = func(restaurantID string) ([]entity.Table, error) {
		return nil, fmt.Errorf("db error")
	}
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	_, err := uc.GetTables("r1")
	if err == nil {
		t.Fatal("expected error from repo failure")
	}
}

func TestRegenerateQR_Success(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", QRCode: "OLDQRCODE"}
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	newCode, err := uc.RegenerateQR("t1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if newCode == "" {
		t.Error("expected non-empty new QR code")
	}
	if newCode == "OLDQRCODE" {
		t.Error("expected new QR code to differ from old one")
	}
	if len(newCode) != 10 {
		t.Errorf("expected QR code of length 10, got %d", len(newCode))
	}
	if tableRepo.Tables["t1"].QRCode != newCode {
		t.Error("expected table QRCode to be updated in repo")
	}
}

func TestRegenerateQR_TableNotFound(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	_, err := uc.RegenerateQR("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent table")
	}
}

func TestRegenerateQR_UpdateError(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", QRCode: "OLDCODE01"}
	tableRepo.UpdateQRCodeFn = func(id, qrCode string) error {
		return fmt.Errorf("db error")
	}
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	_, err := uc.RegenerateQR("t1")
	if err == nil {
		t.Fatal("expected error from update failure")
	}
}
