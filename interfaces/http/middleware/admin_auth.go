package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/infrastructure/auth"
	"go.uber.org/zap"
)

// AdminClaimsKey is the gin context key used to store admin claims.
const AdminClaimsKey = "admin_claims"

// AdminAuthMiddleware validates the Bearer admin JWT and injects claims into the context.
func AdminAuthMiddleware(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No autorizado."})
			return
		}
		claims, err := auth.VerifyAdminSession(strings.TrimPrefix(header, "Bearer "), secret)
		if err != nil {
			zap.L().Warn("invalid admin token",
				zap.String("ip", RealIP(c)),
				zap.Error(err),
			)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Sesión inválida o expirada."})
			return
		}
		c.Set(AdminClaimsKey, claims)
		c.Next()
	}
}

// GetAdminClaims retrieves validated admin claims from the gin context.
func GetAdminClaims(c *gin.Context) *auth.AdminClaims {
	v, _ := c.Get(AdminClaimsKey)
	claims, _ := v.(*auth.AdminClaims)
	return claims
}
