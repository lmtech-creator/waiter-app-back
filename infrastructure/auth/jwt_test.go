package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var testSecret = []byte("test-secret-32bytes-long-enough!!")

func makeClaims(expOffset time.Duration) SessionClaims {
	return SessionClaims{
		TableID:      "t1",
		RestaurantID: "r1",
		TableNumber:  5,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expOffset)),
		},
	}
}

func TestSignSession_Success(t *testing.T) {
	claims := makeClaims(30 * time.Minute)
	token, err := SignSession(claims, testSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token string")
	}
}

func TestVerifySession_Valid(t *testing.T) {
	claims := makeClaims(30 * time.Minute)
	token, _ := SignSession(claims, testSecret)

	got, err := VerifySession(token, testSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.TableID != "t1" {
		t.Errorf("expected TableID t1, got %s", got.TableID)
	}
	if got.RestaurantID != "r1" {
		t.Errorf("expected RestaurantID r1, got %s", got.RestaurantID)
	}
	if got.TableNumber != 5 {
		t.Errorf("expected TableNumber 5, got %d", got.TableNumber)
	}
}

func TestVerifySession_Expired(t *testing.T) {
	claims := makeClaims(-1 * time.Minute)
	token, _ := SignSession(claims, testSecret)

	_, err := VerifySession(token, testSecret)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestVerifySession_WrongAlgorithm(t *testing.T) {
	// HS384 is not in the allowed list (only HS256).
	claims := makeClaims(30 * time.Minute)
	tk := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	signed, _ := tk.SignedString(testSecret)

	_, err := VerifySession(signed, testSecret)
	if err == nil {
		t.Fatal("expected error for disallowed algorithm")
	}
}

func TestVerifySession_WrongSecret(t *testing.T) {
	claims := makeClaims(30 * time.Minute)
	token, _ := SignSession(claims, testSecret)

	_, err := VerifySession(token, []byte("wrong-secret-32bytes-also-enough!"))
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestVerifySession_InvalidTokenString(t *testing.T) {
	_, err := VerifySession("not.a.valid.jwt.at.all", testSecret)
	if err == nil {
		t.Fatal("expected error for invalid token string")
	}
}

func TestVerifySession_NoExpiry(t *testing.T) {
	// WithExpirationRequired must reject a token that has no ExpiresAt.
	claims := SessionClaims{
		TableID:      "t1",
		RestaurantID: "r1",
		TableNumber:  5,
	}
	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := tk.SignedString(testSecret)

	_, err := VerifySession(signed, testSecret)
	if err == nil {
		t.Fatal("expected error for token without expiry")
	}
}

// ─── Admin JWT tests ───────────────────────────────────────────────────────

var adminTestSecret = []byte("admin-test-secret-32bytes-enough")

func makeAdminClaims(expOffset time.Duration) AdminClaims {
	return AdminClaims{
		AdminID:      "admin-1",
		RestaurantID: "r1",
		Role:         "owner",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expOffset)),
		},
	}
}

func TestSignAdminSession_Success(t *testing.T) {
	claims := makeAdminClaims(24 * time.Hour)
	token, err := SignAdminSession(claims, adminTestSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestVerifyAdminSession_Valid(t *testing.T) {
	claims := makeAdminClaims(24 * time.Hour)
	token, _ := SignAdminSession(claims, adminTestSecret)

	got, err := VerifyAdminSession(token, adminTestSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.AdminID != "admin-1" {
		t.Errorf("expected AdminID admin-1, got %s", got.AdminID)
	}
	if got.RestaurantID != "r1" {
		t.Errorf("expected RestaurantID r1, got %s", got.RestaurantID)
	}
	if got.Role != "owner" {
		t.Errorf("expected Role owner, got %s", got.Role)
	}
}

func TestVerifyAdminSession_Expired(t *testing.T) {
	claims := makeAdminClaims(-1 * time.Minute)
	token, _ := SignAdminSession(claims, adminTestSecret)

	_, err := VerifyAdminSession(token, adminTestSecret)
	if err == nil {
		t.Fatal("expected error for expired admin token")
	}
}

func TestVerifyAdminSession_WrongSecret(t *testing.T) {
	claims := makeAdminClaims(24 * time.Hour)
	token, _ := SignAdminSession(claims, adminTestSecret)

	_, err := VerifyAdminSession(token, []byte("wrong-secret-32bytes-also-enough!"))
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestVerifyAdminSession_InvalidToken(t *testing.T) {
	_, err := VerifyAdminSession("not.a.valid.jwt", adminTestSecret)
	if err == nil {
		t.Fatal("expected error for invalid token string")
	}
}

func TestVerifyAdminSession_WrongAlgorithm(t *testing.T) {
	claims := makeAdminClaims(24 * time.Hour)
	tk := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	signed, _ := tk.SignedString(adminTestSecret)

	_, err := VerifyAdminSession(signed, adminTestSecret)
	if err == nil {
		t.Fatal("expected error for disallowed algorithm")
	}
}

func TestVerifyAdminSession_SessionTokenRejected(t *testing.T) {
	// A customer session JWT must not pass admin verification.
	sessionClaims := makeClaims(30 * time.Minute)
	token, _ := SignSession(sessionClaims, adminTestSecret)

	_, err := VerifyAdminSession(token, adminTestSecret)
	if err == nil {
		t.Fatal("expected customer session token to fail admin verification")
	}
}
