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

// ─── UpdateAdminUser ──────────────────────────────────────────────────────────

func TestUpdateAdminUser_SuperAdminUpdatesUsername(t *testing.T) {
	repo := mocks.NewAdminRepo()
	repo.Admins["target"] = &entity.AdminUser{
		ID: "target", Username: "oldname", Role: entity.RoleEmployee,
		RestaurantID: strPtr("r1"),
	}
	uc := NewAdminUseCase(repo, adminTestSecret)

	newUsername := "newname"
	updated, err := uc.UpdateAdminUser(UpdateAdminInput{
		RequesterRole: entity.RoleSuperAdmin,
		TargetID:      "target",
		Username:      &newUsername,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Username != "newname" {
		t.Errorf("expected username 'newname', got '%s'", updated.Username)
	}
}

func TestUpdateAdminUser_SuperAdminUpdatesPassword(t *testing.T) {
	repo := mocks.NewAdminRepo()
	repo.Admins["target"] = &entity.AdminUser{
		ID: "target", Username: "user", PasswordHash: hashPassword(t, "old"),
		Role: entity.RoleEmployee, RestaurantID: strPtr("r1"),
	}
	uc := NewAdminUseCase(repo, adminTestSecret)

	newPw := "newpassword123"
	_, err := uc.UpdateAdminUser(UpdateAdminInput{
		RequesterRole: entity.RoleSuperAdmin,
		TargetID:      "target",
		Password:      &newPw,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	stored := repo.Admins["target"]
	if err := bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte(newPw)); err != nil {
		t.Error("expected stored password to match new password")
	}
}

func TestUpdateAdminUser_OwnerUpdatesOwnEmployee(t *testing.T) {
	repo := mocks.NewAdminRepo()
	rid := "r1"
	repo.Admins["target"] = &entity.AdminUser{
		ID: "target", Username: "emp", Role: entity.RoleEmployee,
		RestaurantID: strPtr(rid),
	}
	uc := NewAdminUseCase(repo, adminTestSecret)

	newUsername := "updated-emp"
	updated, err := uc.UpdateAdminUser(UpdateAdminInput{
		RequesterRole:         entity.RoleOwner,
		RequesterRestaurantID: rid,
		TargetID:              "target",
		Username:              &newUsername,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Username != "updated-emp" {
		t.Errorf("expected 'updated-emp', got '%s'", updated.Username)
	}
}

func TestUpdateAdminUser_OwnerCannotUpdateOtherRestaurant(t *testing.T) {
	repo := mocks.NewAdminRepo()
	repo.Admins["target"] = &entity.AdminUser{
		ID: "target", Username: "emp", Role: entity.RoleEmployee,
		RestaurantID: strPtr("r-other"),
	}
	uc := NewAdminUseCase(repo, adminTestSecret)

	newUsername := "hacker"
	_, err := uc.UpdateAdminUser(UpdateAdminInput{
		RequesterRole:         entity.RoleOwner,
		RequesterRestaurantID: "r1",
		TargetID:              "target",
		Username:              &newUsername,
	})
	if err == nil {
		t.Error("expected error: owner cannot update user of another restaurant")
	}
}

func TestUpdateAdminUser_OwnerCannotUpdateOwner(t *testing.T) {
	repo := mocks.NewAdminRepo()
	rid := "r1"
	repo.Admins["target"] = &entity.AdminUser{
		ID: "target", Username: "owner2", Role: entity.RoleOwner,
		RestaurantID: strPtr(rid),
	}
	uc := NewAdminUseCase(repo, adminTestSecret)

	newUsername := "new"
	_, err := uc.UpdateAdminUser(UpdateAdminInput{
		RequesterRole:         entity.RoleOwner,
		RequesterRestaurantID: rid,
		TargetID:              "target",
		Username:              &newUsername,
	})
	if err == nil {
		t.Error("expected error: owner cannot update another owner")
	}
}

func TestUpdateAdminUser_EmployeeCannotUpdate(t *testing.T) {
	repo := mocks.NewAdminRepo()
	repo.Admins["target"] = &entity.AdminUser{ID: "target", Username: "emp", Role: entity.RoleEmployee}
	uc := NewAdminUseCase(repo, adminTestSecret)

	newUsername := "new"
	_, err := uc.UpdateAdminUser(UpdateAdminInput{
		RequesterRole: entity.RoleEmployee,
		TargetID:      "target",
		Username:      &newUsername,
	})
	if err == nil {
		t.Error("expected error: employee cannot update users")
	}
}

func TestUpdateAdminUser_NotFound(t *testing.T) {
	repo := mocks.NewAdminRepo()
	uc := NewAdminUseCase(repo, adminTestSecret)

	newUsername := "any"
	_, err := uc.UpdateAdminUser(UpdateAdminInput{
		RequesterRole: entity.RoleSuperAdmin,
		TargetID:      "nonexistent",
		Username:      &newUsername,
	})
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestUpdateAdminUser_EmptyUsernameRejected(t *testing.T) {
	repo := mocks.NewAdminRepo()
	repo.Admins["target"] = &entity.AdminUser{ID: "target", Username: "user", Role: entity.RoleEmployee}
	uc := NewAdminUseCase(repo, adminTestSecret)

	empty := ""
	_, err := uc.UpdateAdminUser(UpdateAdminInput{
		RequesterRole: entity.RoleSuperAdmin,
		TargetID:      "target",
		Username:      &empty,
	})
	if err == nil {
		t.Error("expected error for empty username")
	}
}

func TestUpdateAdminUser_EmptyPasswordRejected(t *testing.T) {
	repo := mocks.NewAdminRepo()
	repo.Admins["target"] = &entity.AdminUser{ID: "target", Username: "user", Role: entity.RoleEmployee}
	uc := NewAdminUseCase(repo, adminTestSecret)

	empty := ""
	_, err := uc.UpdateAdminUser(UpdateAdminInput{
		RequesterRole: entity.RoleSuperAdmin,
		TargetID:      "target",
		Password:      &empty,
	})
	if err == nil {
		t.Error("expected error for empty password")
	}
}

// ─── ResetPassword ────────────────────────────────────────────────────────────

func TestResetPassword_SuperAdmin(t *testing.T) {
	repo := mocks.NewAdminRepo()
	repo.Admins["target"] = &entity.AdminUser{
		ID: "target", Username: "user", PasswordHash: hashPassword(t, "old"),
		Role: entity.RoleEmployee,
	}
	uc := NewAdminUseCase(repo, adminTestSecret)

	out, err := uc.ResetPassword(ResetPasswordInput{
		RequesterRole: entity.RoleSuperAdmin,
		TargetID:      "target",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.NewPassword == "" {
		t.Error("expected non-empty new password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repo.Admins["target"].PasswordHash), []byte(out.NewPassword)); err != nil {
		t.Error("expected stored hash to match new password")
	}
}

func TestResetPassword_OwnerResetsOwnEmployee(t *testing.T) {
	repo := mocks.NewAdminRepo()
	rid := "r1"
	repo.Admins["target"] = &entity.AdminUser{
		ID: "target", Username: "emp", Role: entity.RoleEmployee,
		RestaurantID: strPtr(rid),
	}
	uc := NewAdminUseCase(repo, adminTestSecret)

	out, err := uc.ResetPassword(ResetPasswordInput{
		RequesterRole:         entity.RoleOwner,
		RequesterRestaurantID: rid,
		TargetID:              "target",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.NewPassword == "" {
		t.Error("expected non-empty new password")
	}
}

func TestResetPassword_OwnerCannotResetOtherRestaurant(t *testing.T) {
	repo := mocks.NewAdminRepo()
	repo.Admins["target"] = &entity.AdminUser{
		ID: "target", Username: "emp", Role: entity.RoleEmployee,
		RestaurantID: strPtr("r-other"),
	}
	uc := NewAdminUseCase(repo, adminTestSecret)

	_, err := uc.ResetPassword(ResetPasswordInput{
		RequesterRole:         entity.RoleOwner,
		RequesterRestaurantID: "r1",
		TargetID:              "target",
	})
	if err == nil {
		t.Error("expected error: owner cannot reset password of another restaurant")
	}
}

func TestResetPassword_NotFound(t *testing.T) {
	repo := mocks.NewAdminRepo()
	uc := NewAdminUseCase(repo, adminTestSecret)

	_, err := uc.ResetPassword(ResetPasswordInput{
		RequesterRole: entity.RoleSuperAdmin,
		TargetID:      "nonexistent",
	})
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}
