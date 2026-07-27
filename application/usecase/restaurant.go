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
}

func NewRestaurantUseCase(rr repository.RestaurantRepository, tr repository.TableRepository) *RestaurantUseCase {
	return &RestaurantUseCase{restaurantRepo: rr, tableRepo: tr}
}

type CreateRestaurantInput struct {
	Name          string   `json:"name" binding:"required"`
	Plan          string   `json:"plan"`
	QrBannerText  string   `json:"qr_banner_text"`
	QrFooterItems []string `json:"qr_footer_items"`
}

func (uc *RestaurantUseCase) CreateRestaurant(input CreateRestaurantInput) (*entity.Restaurant, error) {
	plan := input.Plan
	if plan == "" {
		plan = "free"
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
		Plan:          plan,
		QrBannerText:  bannerText,
		QrFooterItems: entity.JSONB(footerItems),
	}

	if err := uc.restaurantRepo.Create(r); err != nil {
		return nil, fmt.Errorf("failed to create restaurant: %w", err)
	}

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
	Number       int    `json:"number" binding:"required"`
	RestaurantID string `json:"restaurant_id" binding:"required"`
}

func (uc *RestaurantUseCase) CreateTable(input CreateTableInput) (*entity.Table, error) {
	if _, err := uc.restaurantRepo.FindByID(input.RestaurantID); err != nil {
		return nil, fmt.Errorf("restaurant not found: %w", err)
	}

	if existing, _ := uc.tableRepo.FindByNumberAndRestaurantID(input.Number, input.RestaurantID); existing != nil {
		return nil, ErrTableNumberExists
	}

	qr, err := generateQRCode(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	t := &entity.Table{
		ID:           uuid.New().String(),
		Number:       input.Number,
		RestaurantID: input.RestaurantID,
		QRCode:       qr,
		IsActive:     true,
	}

	if err := uc.tableRepo.Create(t); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return t, nil
}

func (uc *RestaurantUseCase) GetTables(restaurantID string) ([]entity.Table, error) {
	return uc.tableRepo.FindByRestaurantID(restaurantID)
}

func (uc *RestaurantUseCase) RegenerateQR(tableID string) (string, error) {
	if _, err := uc.tableRepo.FindByID(tableID); err != nil {
		return "", fmt.Errorf("table not found: %w", err)
	}

	newCode, err := generateQRCode(10)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %w", err)
	}

	if err := uc.tableRepo.UpdateQRCode(tableID, newCode); err != nil {
		return "", fmt.Errorf("failed to update QR code: %w", err)
	}

	return newCode, nil
}

func (uc *RestaurantUseCase) InactivateTable(restaurantID, tableID string) error {
	t, err := uc.tableRepo.FindByID(tableID)
	if err != nil {
		return ErrTableNotFound
	}
	if t.RestaurantID != restaurantID {
		return ErrTableRestaurantMismatch
	}
	if err := uc.tableRepo.InactiveTable(tableID); err != nil {
		return fmt.Errorf("failed to inactivate table: %w", err)
	}
	return nil
}

func (uc *RestaurantUseCase) ActivateTable(restaurantID, tableID string) error {
	t, err := uc.tableRepo.FindByID(tableID)
	if err != nil {
		return ErrTableNotFound
	}
	if t.RestaurantID != restaurantID {
		return ErrTableRestaurantMismatch
	}
	if err := uc.tableRepo.ReactiVateTable(tableID); err != nil {
		return fmt.Errorf("failed to activate table: %w", err)
	}
	return nil
}
