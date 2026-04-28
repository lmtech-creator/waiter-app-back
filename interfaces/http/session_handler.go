package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/waiter/back/domain/repository"
	"github.com/waiter/back/infrastructure/auth"
	"github.com/waiter/back/interfaces/http/dto"
	mw "github.com/waiter/back/interfaces/http/middleware"
	"go.uber.org/zap"
)

// SessionHandler handles customer session creation via QR code.
type SessionHandler struct {
	tableRepo   repository.TableRepository
	secret      []byte
	signSession func(claims auth.SessionClaims, secret []byte) (string, error)
}

func NewSessionHandler(tableRepo repository.TableRepository, secret []byte) *SessionHandler {
	return &SessionHandler{
		tableRepo:   tableRepo,
		secret:      secret,
		signSession: auth.SignSession,
	}
}

// Create godoc
// @Summary      Iniciar sesión con QR
// @Description  Valida el código QR de una mesa y devuelve un JWT de sesión (30 min)
// @Tags         session
// @Accept       json
// @Produce      json
// @Param        request  body      dto.SessionRequest   true  "Código QR de la mesa"
// @Success      200      {object}  dto.SessionResponse
// @Failure      400      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Router       /api/v1/session [post]
func (h *SessionHandler) Create(c *gin.Context) {
	ip := mw.RealIP(c)

	var req dto.SessionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.QRCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Código QR requerido."})
		return
	}

	table, err := h.tableRepo.FindByQRCode(req.QRCode)
	if err != nil {
		zap.L().Warn("invalid QR attempt",
			zap.String("qr_code", req.QRCode),
			zap.String("ip", ip),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "QR inválido."})
		return
	}

	if !table.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Mesa inactiva."})
		return
	}

	claims := auth.SessionClaims{
		TableID:      table.ID,
		RestaurantID: table.RestaurantID,
		TableNumber:  table.Number,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
		},
	}

	token, err := h.signSession(claims, h.secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno."})
		return
	}

	zap.L().Info("session created",
		zap.Int("table_number", table.Number),
		zap.String("ip", ip),
	)

	c.JSON(http.StatusOK, dto.SessionResponse{
		SessionToken: token,
		Table:        dto.TablePublic{Number: table.Number},
	})
}
