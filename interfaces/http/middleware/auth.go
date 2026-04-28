package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/infrastructure/auth"
	"go.uber.org/zap"
)

// ClaimsKey is the gin context key used to store session claims.
const ClaimsKey = "session_claims"

// AuthMiddleware validates the Bearer JWT and injects claims into the gin context.
func AuthMiddleware(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No autorizado."})
			return
		}
		claims, err := auth.VerifySession(strings.TrimPrefix(header, "Bearer "), secret)
		if err != nil {
			zap.L().Warn("invalid token",
				zap.String("ip", RealIP(c)),
				zap.Error(err),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Sesión inválida o expirada. Escaneá el QR nuevamente."})
			return
		}
		c.Set(ClaimsKey, claims)
		c.Next()
	}
}

// GetClaims retrieves validated session claims from the gin context.
func GetClaims(c *gin.Context) *auth.SessionClaims {
	v, _ := c.Get(ClaimsKey)
	claims, _ := v.(*auth.SessionClaims)
	return claims
}

// RealIP returns the real client IP, respecting X-Forwarded-For.
func RealIP(c *gin.Context) string {
	if fwd := c.GetHeader("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	if ip == "" {
		return c.Request.RemoteAddr
	}
	return ip
}
