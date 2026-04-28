package usecase

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/domain/repository"
	"github.com/waiter/back/infrastructure/auth"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("credenciales incorrectas")

type AdminUseCase struct {
	repo        repository.AdminRepository
	adminSecret []byte
}

func NewAdminUseCase(repo repository.AdminRepository, adminSecret []byte) *AdminUseCase {
	return &AdminUseCase{repo: repo, adminSecret: adminSecret}
}

type AdminLoginInput struct {
	Username string
	Password string
}

type AdminLoginOutput struct {
	Token string
}

func (uc *AdminUseCase) Login(input AdminLoginInput) (*AdminLoginOutput, error) {
	if input.Username == "" || input.Password == "" {
		return nil, ErrInvalidCredentials
	}

	admin, err := uc.repo.FindByUsername(input.Username)
	if err != nil || admin == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	restaurantID := ""
	if admin.RestaurantID != nil {
		restaurantID = *admin.RestaurantID
	}

	token, err := auth.SignAdminSession(auth.AdminClaims{
		AdminID:      admin.ID,
		RestaurantID: restaurantID,
	}, uc.adminSecret)
	if err != nil {
		return nil, fmt.Errorf("error signing token: %w", err)
	}

	return &AdminLoginOutput{Token: token}, nil
}

// SeedAdminIfNeeded creates an "admin" user on first startup if none exists yet.
// The generated password is logged and must be changed immediately in production.
func SeedAdminIfNeeded(repo repository.AdminRepository) (seeded bool, username, password string, err error) {
	exists, err := repo.ExistsAny()
	if err != nil {
		return false, "", "", fmt.Errorf("checking admin existence: %w", err)
	}
	if exists {
		return false, "", "", nil
	}

	rawPassword, err := generatePassword(16)
	if err != nil {
		return false, "", "", fmt.Errorf("generating password: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), 12)
	if err != nil {
		return false, "", "", fmt.Errorf("hashing password: %w", err)
	}

	admin := &entity.AdminUser{
		ID:           uuid.New().String(),
		Username:     "admin",
		PasswordHash: string(hash),
	}

	if err := repo.Create(admin); err != nil {
		return false, "", "", fmt.Errorf("creating admin user: %w", err)
	}

	return true, admin.Username, rawPassword, nil
}

func generatePassword(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := cryptorand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
