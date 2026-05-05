package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/domain/entity"
)

// RequireRole returns a middleware that allows only requests whose admin JWT
// carries one of the specified roles. Must be used after AdminAuthMiddleware.
func RequireRole(roles ...entity.AdminRole) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[string(r)] = true
	}
	return func(c *gin.Context) {
		claims := GetAdminClaims(c)
		if claims == nil || !allowed[claims.Role] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "acceso denegado"})
			return
		}
		c.Next()
	}
}

// RequireRestaurantScope enforces that non-superadmin users can only access
// their own restaurant, matched against the :restaurantId path parameter.
// Must be used after AdminAuthMiddleware.
func RequireRestaurantScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetAdminClaims(c)
		if claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no autorizado"})
			return
		}
		// Superadmin has unrestricted access.
		if claims.Role == string(entity.RoleSuperAdmin) {
			c.Next()
			return
		}
		restaurantID := c.Param("restaurantId")
		if restaurantID != "" && claims.RestaurantID != restaurantID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "acceso denegado"})
			return
		}
		c.Next()
	}
}
