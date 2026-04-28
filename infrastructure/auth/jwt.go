package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SessionClaims holds the JWT payload for a customer session.
type SessionClaims struct {
	TableID      string `json:"mesa_id"`
	RestaurantID string `json:"restaurant_id"`
	TableNumber  int    `json:"table_number"`
	jwt.RegisteredClaims
}

// SignSession creates a signed HS256 JWT token.
func SignSession(claims SessionClaims, secret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// VerifySession parses and validates a JWT, returning the claims.
// Algorithm is restricted to HS256 via WithValidMethods; the key function
// returns the secret unconditionally after that check.
func VerifySession(tokenStr string, secret []byte) (*SessionClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &SessionClaims{},
		func(t *jwt.Token) (interface{}, error) {
			return secret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}
	return token.Claims.(*SessionClaims), nil
}

// AdminClaims holds the JWT payload for an admin session.
type AdminClaims struct {
	AdminID      string `json:"admin_id"`
	RestaurantID string `json:"restaurant_id"`
	jwt.RegisteredClaims
}

// SignAdminSession creates a signed HS256 JWT for an admin.
// If ExpiresAt is not set in the claims, defaults to 24 hours from now.
func SignAdminSession(claims AdminClaims, secret []byte) (string, error) {
	if claims.ExpiresAt == nil {
		claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(24 * time.Hour))
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// VerifyAdminSession parses and validates an admin JWT, returning the claims.
func VerifyAdminSession(tokenStr string, secret []byte) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AdminClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return secret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims := token.Claims.(*AdminClaims)
	if claims.AdminID == "" {
		return nil, fmt.Errorf("invalid token: missing admin_id")
	}
	return claims, nil
}
