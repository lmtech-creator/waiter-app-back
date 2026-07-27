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
var ErrRestaurantIDRequired = errors.New("restaurant_id es requerido para este rol")
var ErrUserNotFound = errors.New("usuario no encontrado")
var ErrForbidden = errors.New("operación no permitida")
var ErrUsernameExists = errors.New("nombre de usuario ya en uso")

type AdminUseCase struct {
	repo        repository.AdminRepository
	adminSecret []byte
}

func NewAdminUseCase(repo repository.AdminRepository, adminSecret []byte) *AdminUseCase {
	return &AdminUseCase{repo: repo, adminSecret: adminSecret}
}

// ─── Login ────────────────────────────────────────────────────────────────────

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
		Role:         string(admin.Role),
	}, uc.adminSecret)
	if err != nil {
		return nil, fmt.Errorf("error signing token: %w", err)
	}

	return &AdminLoginOutput{Token: token}, nil
}

// ─── CreateAdminUser ──────────────────────────────────────────────────────────

type CreateAdminInput struct {
	RequesterRole         entity.AdminRole
	RequesterRestaurantID string // empty for superadmin
	Username              string
	Password              string
	Role                  entity.AdminRole
	RestaurantID          string // target restaurant for owner/employee
}

func (uc *AdminUseCase) CreateAdminUser(input CreateAdminInput) (*entity.AdminUser, error) {
	if input.Username == "" || input.Password == "" {
		return nil, ErrInvalidCredentials
	}

	// Validate requester permissions.
	switch input.RequesterRole {
	case entity.RoleSuperAdmin:
		// can create any role
	case entity.RoleOwner:
		// owners can only create employees for their own restaurant
		if input.Role != entity.RoleEmployee {
			return nil, ErrForbidden
		}
		if input.RestaurantID != input.RequesterRestaurantID {
			return nil, ErrForbidden
		}
	default:
		return nil, ErrForbidden
	}

	// Non-superadmin roles require a restaurant.
	if input.Role != entity.RoleSuperAdmin && input.RestaurantID == "" {
		return nil, ErrRestaurantIDRequired
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	var rid *string
	if input.RestaurantID != "" {
		s := input.RestaurantID
		rid = &s
	}

	admin := &entity.AdminUser{
		ID:           uuid.New().String(),
		RestaurantID: rid,
		Username:     input.Username,
		PasswordHash: string(hash),
		Role:         input.Role,
	}

	if err := uc.repo.Create(admin); err != nil {
		return nil, fmt.Errorf("creating admin user: %w", err)
	}

	return admin, nil
}

// ─── ListAdminUsers ───────────────────────────────────────────────────────────

func (uc *AdminUseCase) ListAdminUsers(requesterRole entity.AdminRole, requesterRestaurantID string) ([]entity.AdminUser, error) {
	if requesterRole == entity.RoleSuperAdmin {
		return uc.repo.FindAll()
	}
	return uc.repo.FindByRestaurantID(requesterRestaurantID)
}

// ─── DeleteAdminUser ──────────────────────────────────────────────────────────

func (uc *AdminUseCase) DeleteAdminUser(requesterRole entity.AdminRole, requesterRestaurantID string, targetID string) error {
	target, err := uc.repo.FindByID(targetID)
	if err != nil || target == nil {
		return ErrUserNotFound
	}

	if err := uc.canModifyUser(requesterRole, requesterRestaurantID, target); err != nil {
		return err
	}

	return uc.repo.DeleteByID(targetID)
}

// ─── UpdateAdminUser ──────────────────────────────────────────────────────────

type UpdateAdminInput struct {
	RequesterRole         entity.AdminRole
	RequesterRestaurantID string
	TargetID              string
	Username              *string // nil = no change
	Password              *string // nil = no change
}

func (uc *AdminUseCase) UpdateAdminUser(input UpdateAdminInput) (*entity.AdminUser, error) {
	target, err := uc.repo.FindByID(input.TargetID)
	if err != nil || target == nil {
		return nil, ErrUserNotFound
	}

	// Permission check (same as Delete).
	if err := uc.canModifyUser(input.RequesterRole, input.RequesterRestaurantID, target); err != nil {
		return nil, err
	}

	if input.Username != nil {
		if *input.Username == "" {
			return nil, fmt.Errorf("username cannot be empty")
		}
		target.Username = *input.Username
	}

	if input.Password != nil {
		if *input.Password == "" {
			return nil, fmt.Errorf("password cannot be empty")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*input.Password), 12)
		if err != nil {
			return nil, fmt.Errorf("hashing password: %w", err)
		}
		target.PasswordHash = string(hash)
	}

	if err := uc.repo.Update(target); err != nil {
		return nil, fmt.Errorf("updating admin user: %w", err)
	}

	return target, nil
}

// ─── ResetPassword ──────────────────────────────────────────────────────────

type ResetPasswordInput struct {
	RequesterRole         entity.AdminRole
	RequesterRestaurantID string
	TargetID              string
}

type ResetPasswordOutput struct {
	NewPassword string
}

func (uc *AdminUseCase) ResetPassword(input ResetPasswordInput) (*ResetPasswordOutput, error) {
	target, err := uc.repo.FindByID(input.TargetID)
	if err != nil || target == nil {
		return nil, ErrUserNotFound
	}

	if err := uc.canModifyUser(input.RequesterRole, input.RequesterRestaurantID, target); err != nil {
		return nil, err
	}

	rawPassword, err := generatePassword(16)
	if err != nil {
		return nil, fmt.Errorf("generating password: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), 12)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	target.PasswordHash = string(hash)

	if err := uc.repo.Update(target); err != nil {
		return nil, fmt.Errorf("updating admin user: %w", err)
	}

	return &ResetPasswordOutput{NewPassword: rawPassword}, nil
}

// canModifyUser checks whether the requester has permission to modify/delete the target user.
func (uc *AdminUseCase) canModifyUser(requesterRole entity.AdminRole, requesterRestaurantID string, target *entity.AdminUser) error {
	switch requesterRole {
	case entity.RoleSuperAdmin:
		return nil
	case entity.RoleOwner:
		if target.Role != entity.RoleEmployee {
			return ErrForbidden
		}
		if target.RestaurantID == nil || *target.RestaurantID != requesterRestaurantID {
			return ErrForbidden
		}
		return nil
	default:
		return ErrForbidden
	}
}

// ─── SeedAdminIfNeeded ────────────────────────────────────────────────────────

// SeedAdminIfNeeded creates a superadmin on first startup if none exists yet.
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
		RestaurantID: nil,
		Username:     "admin",
		PasswordHash: string(hash),
		Role:         entity.RoleSuperAdmin,
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
