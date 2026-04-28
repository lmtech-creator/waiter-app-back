package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/application/usecase"
	"github.com/waiter/back/interfaces/http/dto"
	"go.uber.org/zap"
)

type AdminHandler struct {
	uc *usecase.AdminUseCase
}

func NewAdminHandler(uc *usecase.AdminUseCase) *AdminHandler {
	return &AdminHandler{uc: uc}
}

// Login godoc
// @Summary      Login de administrador
// @Description  Valida credenciales y retorna un JWT de administrador con duración de 24 horas
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request  body      dto.AdminLoginRequest   true  "Credenciales de administrador"
// @Success      200      {object}  dto.AdminLoginResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      401      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/v1/admin/login [post]
func (h *AdminHandler) Login(c *gin.Context) {
	var req dto.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: "Cuerpo inválido."})
		return
	}

	out, err := h.uc.Login(usecase.AdminLoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		zap.L().Warn("admin login failed",
			zap.String("username", req.Username),
			zap.String("ip", c.ClientIP()),
		)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: "Credenciales incorrectas."})
		return
	}

	zap.L().Info("admin login success",
		zap.String("username", req.Username),
		zap.String("ip", c.ClientIP()),
	)
	c.JSON(http.StatusOK, dto.AdminLoginResponse{Token: out.Token})
}
