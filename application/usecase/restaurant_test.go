package usecase

import (
	"errors"
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

func TestCreateRestaurant_QRDefaults(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	r, err := uc.CreateRestaurant(CreateRestaurantInput{Name: "Test"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if r.QrBannerText != "Escaneá y llamá al mozo" {
		t.Errorf("expected default banner, got %q", r.QrBannerText)
	}
	if len(r.QrFooterItems) != 3 {
		t.Fatalf("expected 3 default footer items, got %d", len(r.QrFooterItems))
	}
	if r.QrFooterItems[0] != "Llamar al mozo" {
		t.Errorf("expected first footer item 'Llamar al mozo', got %q", r.QrFooterItems[0])
	}
}

func TestCreateRestaurant_WithCustomQRFields(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	r, err := uc.CreateRestaurant(CreateRestaurantInput{
		Name:          "Custom",
		QrBannerText:  "Bienvenido!",
		QrFooterItems: []string{"Llamar", "Reseña"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if r.QrBannerText != "Bienvenido!" {
		t.Errorf("expected 'Bienvenido!', got %q", r.QrBannerText)
	}
	if len(r.QrFooterItems) != 2 {
		t.Fatalf("expected 2 footer items, got %d", len(r.QrFooterItems))
	}
}

func TestCreateRestaurant_InvalidQRBannerText_TooLong(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	_, err := uc.CreateRestaurant(CreateRestaurantInput{
		Name:         "Test",
		QrBannerText: string(make([]byte, 101)),
	})
	if !errors.Is(err, ErrInvalidQRBannerText) {
		t.Fatalf("expected ErrInvalidQRBannerText, got %v", err)
	}
}

func TestCreateRestaurant_InvalidQRFooterItems_TooMany(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	_, err := uc.CreateRestaurant(CreateRestaurantInput{
		Name:          "Test",
		QrFooterItems: []string{"a", "b", "c", "d", "e", "f"},
	})
	if !errors.Is(err, ErrInvalidQRFooterItems) {
		t.Fatalf("expected ErrInvalidQRFooterItems, got %v", err)
	}
}

func TestCreateRestaurant_InvalidQRFooterItems_ItemTooLong(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	_, err := uc.CreateRestaurant(CreateRestaurantInput{
		Name:          "Test",
		QrFooterItems: []string{string(make([]byte, 51))},
	})
	if !errors.Is(err, ErrInvalidQRFooterItems) {
		t.Fatalf("expected ErrInvalidQRFooterItems, got %v", err)
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

func TestInactivateTable_Success(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", RestaurantID: "r1", IsActive: true}
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	err := uc.InactivateTable("r1", "t1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tableRepo.Tables["t1"].IsActive {
		t.Fatal("expected table to be inactive")
	}
}

func TestInactivateTable_NotFound(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	err := uc.InactivateTable("r1", "missing")
	if !errors.Is(err, ErrTableNotFound) {
		t.Fatalf("expected ErrTableNotFound, got %v", err)
	}
}

func TestInactivateTable_RestaurantMismatch(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", RestaurantID: "r2", IsActive: true}
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	err := uc.InactivateTable("r1", "t1")
	if !errors.Is(err, ErrTableRestaurantMismatch) {
		t.Fatalf("expected ErrTableRestaurantMismatch, got %v", err)
	}
}

func TestActivateTable_Success(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", RestaurantID: "r1", IsActive: false}
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	err := uc.ActivateTable("r1", "t1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !tableRepo.Tables["t1"].IsActive {
		t.Fatal("expected table to be active")
	}
}

func TestActivateTable_RepoError(t *testing.T) {
	restaurantRepo := mocks.NewRestaurantRepo()
	tableRepo := mocks.NewTableRepo()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", RestaurantID: "r1", IsActive: false}
	tableRepo.ReactiVateTableFn = func(id string) error {
		return fmt.Errorf("db error")
	}
	uc := NewRestaurantUseCase(restaurantRepo, tableRepo)

	err := uc.ActivateTable("r1", "t1")
	if err == nil {
		t.Fatal("expected error")
	}
}
