package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/infrastructure/auth"
	ws "github.com/waiter/back/infrastructure/websocket"
	"go.uber.org/zap"
)

type WSHandler struct {
	hub         *ws.Hub
	secret      []byte
	adminSecret []byte
}

func NewWSHandler(hub *ws.Hub, secret []byte, adminSecret []byte) *WSHandler {
	return &WSHandler{hub: hub, secret: secret, adminSecret: adminSecret}
}

// Connect godoc
// @Summary      Conectar al WebSocket del restaurante
// @Description  Acepta JWT de admin (admin_secret) o de sesión de cliente. Requiere ?token=<jwt>
// @Tags         websocket
// @Param        restaurantId  path   string  true  "ID del restaurante"
// @Param        token         query  string  true  "JWT de admin o de sesión de cliente"
// @Success      101
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/ws/{restaurantId} [get]
func (h *WSHandler) Connect(c *gin.Context) {
	token := c.Query("token")
	restaurantID := c.Param("restaurantId")

	if token == "" {
		zap.L().Warn("ws auth failed: no token", zap.String("ip", c.ClientIP()), zap.String("restaurant_id", restaurantID))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token requerido"})
		return
	}

	// Try admin JWT first
	if adminClaims, err := auth.VerifyAdminSession(token, h.adminSecret); err == nil {
		if adminClaims.RestaurantID != restaurantID {
			zap.L().Warn("ws admin restaurant mismatch",
				zap.String("ip", c.ClientIP()),
				zap.String("claimed_rid", adminClaims.RestaurantID),
			)
			c.JSON(http.StatusForbidden, gin.H{"error": "acceso denegado"})
			return
		}
		h.hub.Subscribe(c.Writer, c.Request, restaurantID)
		return
	}

	// Fall back to customer session JWT
	claims, err := auth.VerifySession(token, h.secret)
	if err != nil {
		zap.L().Warn("ws auth failed", zap.String("ip", c.ClientIP()), zap.String("restaurant_id", restaurantID), zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token inválido o expirado"})
		return
	}

	if claims.RestaurantID != restaurantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "acceso denegado"})
		return
	}

	h.hub.Subscribe(c.Writer, c.Request, restaurantID)
}
