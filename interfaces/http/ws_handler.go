package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/infrastructure/auth"
	ws "github.com/waiter/back/infrastructure/websocket"
	"go.uber.org/zap"
)

type WSHandler struct {
	hub    *ws.Hub
	secret []byte
}

func NewWSHandler(hub *ws.Hub, secret []byte) *WSHandler {
	return &WSHandler{hub: hub, secret: secret}
}

// Connect godoc
// @Summary      Conectar al WebSocket del restaurante
// @Description  Abre una conexión WebSocket para recibir eventos en tiempo real. Requiere ?token=<jwt>
// @Tags         websocket
// @Param        restaurantId  path   string  true  "ID del restaurante"
// @Param        token         query  string  true  "JWT de sesión válido para ese restaurante"
// @Router       /api/v1/ws/{restaurantId} [get]
func (h *WSHandler) Connect(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token requerido"})
		return
	}

	claims, err := auth.VerifySession(token, h.secret)
	if err != nil {
		zap.L().Warn("ws auth failed", zap.String("ip", c.ClientIP()), zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token inválido o expirado"})
		return
	}

	restaurantID := c.Param("restaurantId")
	if claims.RestaurantID != restaurantID {
		c.JSON(http.StatusForbidden, gin.H{"error": "acceso denegado"})
		return
	}

	h.hub.Subscribe(c.Writer, c.Request, restaurantID)
}
