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

func TestAdminLogin_Success(t *testing.T) {
	repo := mocks.NewAdminRepo()
	rid := uuid.New().String()
	repo.Admins["a1"] = &entity.AdminUser{
		ID:           "a1",
		RestaurantID: &rid,
		Username:     "admin",
		PasswordHash: hashPassword(t, "password123"),
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

func TestSeedAdminIfNeeded_CreatesAdmin(t *testing.T) {
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
}

func TestSeedAdminIfNeeded_SkipsIfExists(t *testing.T) {
	repo := mocks.NewAdminRepo()
	repo.Admins["existing"] = &entity.AdminUser{ID: "existing", Username: "admin"}

	seeded, _, _, err := SeedAdminIfNeeded(repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seeded {
		t.Error("expected seed to be skipped when admin already exists")
	}
}
