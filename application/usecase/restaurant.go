package usecase

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/domain/repository"
)

var ErrTableNumberExists = errors.New("table number already exists for this restaurant")
var ErrTableNotFound = errors.New("table not found")
var ErrTableRestaurantMismatch = errors.New("table does not belong to restaurant")
var ErrInvalidQRBannerText = errors.New("qr_banner_text must be 1-100 characters")
var ErrInvalidQRFooterItems = errors.New("qr_footer_items must have 1-5 items, each 1-50 characters")

const qrAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateQRCode(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = qrAlphabet[int(b[i])%len(qrAlphabet)]
	}
	return string(b), nil
}

type RestaurantUseCase struct {
	restaurantRepo repository.RestaurantRepository
	tableRepo      repository.TableRepository
	planRepo       repository.PlanRepository
}

func NewRestaurantUseCase(
	rr repository.RestaurantRepository,
	tr repository.TableRepository,
	pr repository.PlanRepository,
) *RestaurantUseCase {
	return &RestaurantUseCase{restaurantRepo: rr, tableRepo: tr, planRepo: pr}
}

type CreateRestaurantInput struct {
	Name          string   `json:"name" binding:"required"`
	PlanID        string   `json:"plan_id"`
	Plan          string   `json:"plan"`
	QrBannerText  string   `json:"qr_banner_text"`
	QrFooterItems []string `json:"qr_footer_items"`
}

func (uc *RestaurantUseCase) resolvePlan(planID, planName string) (*entity.Plan, error) {
	if planID != "" {
		plan, err := uc.planRepo.FindByID(planID)
		if err != nil {
			return nil, ErrPlanNotFound
		}
		return plan, nil
	}
	if planName != "" {
		plan, err := uc.planRepo.FindByName(planName)
		if err != nil {
			return nil, ErrPlanNotFound
		}
		return plan, nil
	}
	plan, err := uc.planRepo.FindByName("free")
	if err != nil {
		return nil, ErrPlanNotFound
	}
	return plan, nil
}

func (uc *RestaurantUseCase) CreateRestaurant(input CreateRestaurantInput) (*entity.Restaurant, error) {
	plan, err := uc.resolvePlan(input.PlanID, input.Plan)
	if err != nil {
		return nil, err
	}

	bannerText := input.QrBannerText
	if bannerText == "" {
		bannerText = "Escaneá y llamá al mozo"
	}
	if len(bannerText) > 100 {
		return nil, ErrInvalidQRBannerText
	}

	footerItems := input.QrFooterItems
	if footerItems == nil {
		footerItems = []string{"Llamar al mozo", "Pedir la cuenta", "Dejar reseña"}
	}
	if err := validateFooterItems(footerItems); err != nil {
		return nil, err
	}

	r := &entity.Restaurant{
		ID:            uuid.New().String(),
		Name:          input.Name,
		PlanID:        plan.ID,
		QrBannerText:  bannerText,
		QrFooterItems: entity.JSONB(footerItems),
	}
	if err := uc.restaurantRepo.Create(r); err != nil {
		return nil, fmt.Errorf("failed to create restaurant: %w", err)
	}
	// Populate the association for the response only, after persisting, so
	// GORM does not attempt to upsert the plan row on create.
	r.Plan = *plan
	return r, nil
}

func validateFooterItems(items []string) error {
	if len(items) < 1 || len(items) > 5 {
		return ErrInvalidQRFooterItems
	}
	for _, item := range items {
		if len(item) < 1 || len(item) > 50 {
			return ErrInvalidQRFooterItems
		}
	}
	return nil
}

func (uc *RestaurantUseCase) GetRestaurant(id string) (*entity.Restaurant, error) {
	return uc.restaurantRepo.FindByID(id)
}

func (uc *RestaurantUseCase) GetAllRestaurants() ([]entity.Restaurant, error) {
	return uc.restaurantRepo.FindAll()
}

type CreateTableInput struct {
	Number       int    `json:"number"`
	RestaurantID string `json:"restaurant_id"`
}

func (uc *RestaurantUseCase) CreateTable(input CreateTableInput) (*entity.Table, error) {
	if _, err := uc.restaurantRepo.FindByID(input.RestaurantID); err != nil {
		return nil, fmt.Errorf("restaurant not found: %w", err)
	}

	existing, err := uc.tableRepo.FindByNumberAndRestaurantID(input.Number, input.RestaurantID)
	if err == nil && existing != nil {
		return nil, ErrTableNumberExists
	}

	qrCode, err := generateQRCode(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	table := &entity.Table{
		ID:           uuid.New().String(),
		Number:       input.Number,
		RestaurantID: input.RestaurantID,
		QRCode:       qrCode,
		IsActive:     true,
	}
	if err := uc.tableRepo.Create(table); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	return table, nil
}

func (uc *RestaurantUseCase) GetTables(restaurantID string) ([]entity.Table, error) {
	return uc.tableRepo.FindByRestaurantID(restaurantID)
}

func (uc *RestaurantUseCase) RegenerateQR(tableID string) (string, error) {
	if _, err := uc.tableRepo.FindByID(tableID); err != nil {
		return "", fmt.Errorf("table not found: %w", err)
	}

	qrCode, err := generateQRCode(10)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %w", err)
	}
	if err := uc.tableRepo.UpdateQRCode(tableID, qrCode); err != nil {
		return "", fmt.Errorf("failed to update QR code: %w", err)
	}

	return qrCode, nil
}

func (uc *RestaurantUseCase) InactivateTable(restaurantID, tableID string) error {
	table, err := uc.tableRepo.FindByID(tableID)
	if err != nil {
		return ErrTableNotFound
	}
	if table.RestaurantID != restaurantID {
		return ErrTableRestaurantMismatch
	}
	return uc.tableRepo.InactiveTable(tableID)
}

func (uc *RestaurantUseCase) ActivateTable(restaurantID, tableID string) error {
	table, err := uc.tableRepo.FindByID(tableID)
	if err != nil {
		return ErrTableNotFound
	}
	if table.RestaurantID != restaurantID {
		return ErrTableRestaurantMismatch
	}
	return uc.tableRepo.ReactiVateTable(tableID)
}

func (uc *RestaurantUseCase) DeleteTable(restaurantID, tableID string) error {
	table, err := uc.tableRepo.FindByID(tableID)
	if err != nil {
		return ErrTableNotFound
	}
	if table.RestaurantID != restaurantID {
		return ErrTableRestaurantMismatch
	}
	return uc.tableRepo.DeleteByID(tableID)
}
