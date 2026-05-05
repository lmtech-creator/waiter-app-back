package usecase

import (
	"testing"

	"github.com/google/uuid"
	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/mocks"
	"golang.org/x/crypto/bcrypt"
)

var adminTestSecret = []byte("admin-test-secret-32bytes-enough")

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	return string(h)
}

func strPtr(s string) *string { return &s }

func TestAdminLogin_Success(t *testing.T) {
	repo := mocks.NewAdminRepo()
	rid := uuid.New().String()
	repo.Admins["a1"] = &entity.AdminUser{
		ID:           "a1",
		RestaurantID: strPtr(rid),
		Username:     "admin",
		PasswordHash: hashPassword(t, "password123"),
		Role:         entity.RoleOwner,
	}

	uc := NewAdminUseCase(repo, adminTestSecret)
	out, err := uc.Login(AdminLoginInput{Username: "admin", Password: "password123"})
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if out.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestAdminLogin_WrongPassword(t *testing.T) {
	repo := mocks.NewAdminRepo()
	repo.Admins["a1"] = &entity.AdminUser{
		ID:           "a1",
		Username:     "admin",
		PasswordHash: hashPassword(t, "correct"),
		Role:         entity.RoleOwner,
	}

	uc := NewAdminUseCase(repo, adminTestSecret)
	_, err := uc.Login(AdminLoginInput{Username: "admin", Password: "wrong"})
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestAdminLogin_UnknownUser(t *testing.T) {
	repo := mocks.NewAdminRepo()
	uc := NewAdminUseCase(repo, adminTestSecret)
	_, err := uc.Login(AdminLoginInput{Username: "nobody", Password: "x"})
	if err == nil {
		t.Fatal("expected error for unknown user")
	}
}

func TestAdminLogin_EmptyFields(t *testing.T) {
	repo := mocks.NewAdminRepo()
	uc := NewAdminUseCase(repo, adminTestSecret)

	if _, err := uc.Login(AdminLoginInput{Username: "", Password: "x"}); err == nil {
		t.Error("expected error for empty username")
	}
	if _, err := uc.Login(AdminLoginInput{Username: "admin", Password: ""}); err == nil {
		t.Error("expected error for empty password")
	}
}

func TestSeedAdminIfNeeded_CreatesSuperAdmin(t *testing.T) {
	repo := mocks.NewAdminRepo()
	seeded, username, password, err := SeedAdminIfNeeded(repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !seeded {
		t.Fatal("expected admin to be seeded")
	}
	if username == "" || password == "" {
		t.Error("expected non-empty username and password")
	}
	if len(repo.Admins) != 1 {
		t.Errorf("expected 1 admin in repo, got %d", len(repo.Admins))
	}
	for _, a := range repo.Admins {
		if a.Role != entity.RoleSuperAdmin {
			t.Errorf("expected role superadmin, got %s", a.Role)
		}
		if a.RestaurantID != nil {
			t.Error("expected nil restaurant_id for superadmin")
		}
	}
}

func TestSeedAdminIfNeeded_SkipsIfExists(t *testing.T) {
	repo := mocks.NewAdminRepo()
	repo.Admins["existing"] = &entity.AdminUser{ID: "existing", Username: "admin", Role: entity.RoleSuperAdmin}

	seeded, _, _, err := SeedAdminIfNeeded(repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seeded {
		t.Error("expected seed to be skipped when admin already exists")
	}
}

func TestCreateAdminUser_SuperAdminCanCreateOwner(t *testing.T) {
	repo := mocks.NewAdminRepo()
	uc := NewAdminUseCase(repo, adminTestSecret)
	rid := uuid.New().String()

	admin, err := uc.CreateAdminUser(CreateAdminInput{
		RequesterRole: entity.RoleSuperAdmin,
		Username:      "owner1",
		Password:      "pass123",
		Role:          entity.RoleOwner,
		RestaurantID:  rid,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if admin.Role != entity.RoleOwner {
		t.Errorf("expected role owner, got %s", admin.Role)
	}
}

func TestCreateAdminUser_OwnerCanCreateEmployee(t *testing.T) {
	repo := mocks.NewAdminRepo()
	uc := NewAdminUseCase(repo, adminTestSecret)
	rid := "restaurant-a"

	admin, err := uc.CreateAdminUser(CreateAdminInput{
		RequesterRole:         entity.RoleOwner,
		RequesterRestaurantID: rid,
		Username:              "emp1",
		Password:              "pass123",
		Role:                  entity.RoleEmployee,
		RestaurantID:          rid,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if admin.Role != entity.RoleEmployee {
		t.Errorf("expected role employee, got %s", admin.Role)
	}
}

func TestCreateAdminUser_OwnerCannotCreateOwner(t *testing.T) {
	repo := mocks.NewAdminRepo()
	uc := NewAdminUseCase(repo, adminTestSecret)
	rid := "restaurant-a"

	_, err := uc.CreateAdminUser(CreateAdminInput{
		RequesterRole:         entity.RoleOwner,
		RequesterRestaurantID: rid,
		Username:              "owner2",
		Password:              "pass123",
		Role:                  entity.RoleOwner,
		RestaurantID:          rid,
	})
	if err == nil {
		t.Error("expected error: owner cannot create owner")
	}
}

func TestCreateAdminUser_OwnerCannotCreateForOtherRestaurant(t *testing.T) {
	repo := mocks.NewAdminRepo()
	uc := NewAdminUseCase(repo, adminTestSecret)

	_, err := uc.CreateAdminUser(CreateAdminInput{
		RequesterRole:         entity.RoleOwner,
		RequesterRestaurantID: "restaurant-a",
		Username:              "emp1",
		Password:              "pass123",
		Role:                  entity.RoleEmployee,
		RestaurantID:          "restaurant-b",
	})
	if err == nil {
		t.Error("expected error: owner cannot create users for another restaurant")
	}
}

func TestDeleteAdminUser_SuperAdminCanDeleteAny(t *testing.T) {
	repo := mocks.NewAdminRepo()
	rid := strPtr("restaurant-a")
	repo.Admins["target"] = &entity.AdminUser{ID: "target", Username: "emp", Role: entity.RoleEmployee, RestaurantID: rid}

	uc := NewAdminUseCase(repo, adminTestSecret)
	err := uc.DeleteAdminUser(entity.RoleSuperAdmin, "", "target")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := repo.Admins["target"]; ok {
		t.Error("expected user to be deleted")
	}
}

func TestDeleteAdminUser_OwnerCanDeleteEmployee(t *testing.T) {
	repo := mocks.NewAdminRepo()
	rid := strPtr("restaurant-a")
	repo.Admins["target"] = &entity.AdminUser{ID: "target", Username: "emp", Role: entity.RoleEmployee, RestaurantID: rid}

	uc := NewAdminUseCase(repo, adminTestSecret)
	err := uc.DeleteAdminUser(entity.RoleOwner, "restaurant-a", "target")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteAdminUser_OwnerCannotDeleteOwner(t *testing.T) {
	repo := mocks.NewAdminRepo()
	rid := strPtr("restaurant-a")
	repo.Admins["target"] = &entity.AdminUser{ID: "target", Username: "owner2", Role: entity.RoleOwner, RestaurantID: rid}

	uc := NewAdminUseCase(repo, adminTestSecret)
	err := uc.DeleteAdminUser(entity.RoleOwner, "restaurant-a", "target")
	if err == nil {
		t.Error("expected error: owner cannot delete another owner")
	}
}

func TestDeleteAdminUser_NotFound(t *testing.T) {
	repo := mocks.NewAdminRepo()
	uc := NewAdminUseCase(repo, adminTestSecret)
	err := uc.DeleteAdminUser(entity.RoleSuperAdmin, "", "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}
