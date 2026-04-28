package auth

import (
	"fmt"

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
func VerifySession(tokenStr string, secret []byte) (*SessionClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &SessionClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return secret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return token.Claims.(*SessionClaims), nil
}
