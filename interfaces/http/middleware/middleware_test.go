package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/waiter/back/infrastructure/auth"
	"golang.org/x/time/rate"
)

func init() {
	gin.SetMode(gin.TestMode)
}

var mwTestSecret = []byte("test-secret-32bytes-middleware!!")

func makeTestToken(expOffset time.Duration) string {
	claims := auth.SessionClaims{
		TableID:      "t1",
		RestaurantID: "r1",
		TableNumber:  1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expOffset)),
		},
	}
	token, _ := auth.SignSession(claims, mwTestSecret)
	return token
}

// ─── AuthMiddleware ───────────────────────────────────────────────────────────

func TestAuthMiddleware_NoHeader(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	AuthMiddleware(mwTestSecret)(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_NoBearerPrefix(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "Token abc123")

	AuthMiddleware(mwTestSecret)(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid.token.here")

	AuthMiddleware(mwTestSecret)(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	token := makeTestToken(30 * time.Minute)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	AuthMiddleware(mwTestSecret)(c)

	if w.Code == http.StatusUnauthorized {
		t.Error("expected valid request to pass middleware")
	}
	claims := GetClaims(c)
	if claims == nil {
		t.Fatal("expected claims to be set in context")
	}
	if claims.TableID != "t1" {
		t.Errorf("expected TableID t1, got %s", claims.TableID)
	}
}

// ─── GetClaims ────────────────────────────────────────────────────────────────

func TestGetClaims_Present(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	expected := &auth.SessionClaims{TableID: "t1"}
	c.Set(ClaimsKey, expected)

	got := GetClaims(c)
	if got != expected {
		t.Error("expected to get the claims that were set")
	}
}

func TestGetClaims_Absent(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	got := GetClaims(c)
	if got != nil {
		t.Errorf("expected nil claims, got %v", got)
	}
}

// ─── RealIP ───────────────────────────────────────────────────────────────────

func TestRealIP_XForwardedFor(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")

	ip := RealIP(c)
	if ip != "1.2.3.4" {
		t.Errorf("expected 1.2.3.4, got %s", ip)
	}
}

func TestRealIP_RemoteAddr(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.RemoteAddr = "10.0.0.1:54321"

	ip := RealIP(c)
	if ip != "10.0.0.1" {
		t.Errorf("expected 10.0.0.1, got %s", ip)
	}
}

func TestRealIP_RemoteAddrNoPort(t *testing.T) {
	// SplitHostPort fails → fallback to raw RemoteAddr.
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.RemoteAddr = "10.0.0.2"

	ip := RealIP(c)
	if ip != "10.0.0.2" {
		t.Errorf("expected 10.0.0.2, got %s", ip)
	}
}

// ─── RateLimitMiddleware ──────────────────────────────────────────────────────

func TestRateLimitMiddleware_Allow(t *testing.T) {
	mwFn := RateLimitMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.RemoteAddr = "192.168.250.10:1234"

	mwFn(c)

	if w.Code == http.StatusTooManyRequests {
		t.Error("expected first request to be allowed")
	}
}

func TestRateLimitMiddleware_ExceedsLimit(t *testing.T) {
	mwFn := RateLimitMiddleware()
	testAddr := "192.168.250.20:9999"
	var lastCode int

	// Burst is 20; 21st request must be rejected.
	for i := 0; i < 21; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Request.RemoteAddr = testAddr
		mwFn(c)
		lastCode = w.Code
	}

	if lastCode != http.StatusTooManyRequests {
		t.Errorf("expected 429 after exceeding burst, got %d", lastCode)
	}
}

// ─── purgeStaleLimiters ───────────────────────────────────────────────────────

func TestPurgeStaleLimiters(t *testing.T) {
	staleIP := "192.168.200.1"
	freshIP := "192.168.200.2"

	mu.Lock()
	limiters[staleIP] = &ipLimiter{
		limiter:  rate.NewLimiter(rate.Every(time.Minute), 60),
		lastSeen: time.Now().Add(-11 * time.Minute), // stale
	}
	limiters[freshIP] = &ipLimiter{
		limiter:  rate.NewLimiter(rate.Every(time.Minute), 60),
		lastSeen: time.Now(), // fresh
	}
	mu.Unlock()

	purgeStaleLimiters()

	mu.Lock()
	_, staleExists := limiters[staleIP]
	_, freshExists := limiters[freshIP]
	delete(limiters, freshIP) // cleanup
	mu.Unlock()

	if staleExists {
		t.Error("expected stale IP to be purged")
	}
	if !freshExists {
		t.Error("expected fresh IP to remain")
	}
}

// ─── AdminAuthMiddleware ──────────────────────────────────────────────────────

var mwAdminSecret = []byte("admin-test-secret-32bytes-enough")

func makeAdminTestToken(expOffset time.Duration) string {
	claims := auth.AdminClaims{
		AdminID:      "a1",
		RestaurantID: "r1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expOffset)),
		},
	}
	token, _ := auth.SignAdminSession(claims, mwAdminSecret)
	return token
}

func TestAdminAuthMiddleware_NoHeader(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	AdminAuthMiddleware(mwAdminSecret)(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAdminAuthMiddleware_InvalidToken(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid.token")

	AdminAuthMiddleware(mwAdminSecret)(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAdminAuthMiddleware_ExpiredToken(t *testing.T) {
	token := makeAdminTestToken(-1 * time.Minute)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	AdminAuthMiddleware(mwAdminSecret)(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAdminAuthMiddleware_ValidToken(t *testing.T) {
	token := makeAdminTestToken(24 * time.Hour)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	AdminAuthMiddleware(mwAdminSecret)(c)

	if w.Code == http.StatusUnauthorized {
		t.Error("expected valid admin token to pass middleware")
	}
	claims := GetAdminClaims(c)
	if claims == nil {
		t.Fatal("expected admin claims to be set in context")
	}
	if claims.AdminID != "a1" {
		t.Errorf("expected AdminID a1, got %s", claims.AdminID)
	}
}

func TestAdminAuthMiddleware_CustomerTokenRejected(t *testing.T) {
	// A customer session JWT signed with the admin secret must be rejected.
	sessionClaims := auth.SessionClaims{
		TableID:      "t1",
		RestaurantID: "r1",
		TableNumber:  1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
		},
	}
	token, _ := auth.SignSession(sessionClaims, mwAdminSecret)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	AdminAuthMiddleware(mwAdminSecret)(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for customer token on admin middleware, got %d", w.Code)
	}
}
