package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/application/usecase"
	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/interfaces/http/dto"
	mw "github.com/waiter/back/interfaces/http/middleware"
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

// CreateUser godoc
// @Summary      Crear usuario administrador
// @Description  Crea un nuevo usuario owner o employee. Superadmin puede crear cualquier rol; owner solo puede crear employees de su restaurante.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateAdminUserRequest  true  "Datos del nuevo usuario"
// @Success      201      {object}  dto.AdminUserResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      403      {object}  dto.ErrorResponse
// @Failure      422      {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/admin/users [post]
func (h *AdminHandler) CreateUser(c *gin.Context) {
	claims := mw.GetAdminClaims(c)

	var req dto.CreateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	requesterRestaurantID := ""
	if claims != nil {
		requesterRestaurantID = claims.RestaurantID
	}

	admin, err := h.uc.CreateAdminUser(usecase.CreateAdminInput{
		RequesterRole:         entity.AdminRole(claims.Role),
		RequesterRestaurantID: requesterRestaurantID,
		Username:              req.Username,
		Password:              req.Password,
		Role:                  entity.AdminRole(req.Role),
		RestaurantID:          req.RestaurantID,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrForbidden):
			c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "forbidden", Message: "Operación no permitida."})
		case errors.Is(err, usecase.ErrRestaurantIDRequired):
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: err.Error()})
		case errors.Is(err, usecase.ErrInvalidCredentials):
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: "Usuario y contraseña requeridos."})
		default:
			c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: "unprocessable_entity", Message: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.ToAdminUserResponse(admin))
}

// ListUsers godoc
// @Summary      Listar usuarios administradores
// @Description  Superadmin ve todos los usuarios; owner ve solo los employees de su restaurante.
// @Tags         admin
// @Produce      json
// @Success      200  {array}   dto.AdminUserResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	claims := mw.GetAdminClaims(c)

	users, err := h.uc.ListAdminUsers(entity.AdminRole(claims.Role), claims.RestaurantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToAdminUserListResponse(users))
}

// DeleteUser godoc
// @Summary      Eliminar usuario administrador
// @Description  Superadmin puede eliminar cualquier usuario; owner solo puede eliminar employees de su restaurante.
// @Tags         admin
// @Produce      json
// @Param        id   path      string  true  "ID del usuario"
// @Success      204
// @Failure      403  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/admin/users/{id} [delete]
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	claims := mw.GetAdminClaims(c)
	targetID := c.Param("id")

	if err := h.uc.DeleteAdminUser(entity.AdminRole(claims.Role), claims.RestaurantID, targetID); err != nil {
		switch {
		case errors.Is(err, usecase.ErrForbidden):
			c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "forbidden", Message: "Operación no permitida."})
		case errors.Is(err, usecase.ErrUserNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "Usuario no encontrado."})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateUser godoc
// @Summary      Modificar usuario administrador
// @Description  Actualiza nombre de usuario y/o contraseña. Superadmin puede modificar cualquier usuario; owner solo employees de su restaurante.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id       path    string                      true  "ID del usuario"
// @Param        request  body    dto.UpdateAdminUserRequest   true  "Campos a modificar"
// @Success      200      {object}  dto.AdminUserResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      403      {object}  dto.ErrorResponse
// @Failure      404      {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/admin/users/{id} [patch]
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	claims := mw.GetAdminClaims(c)
	targetID := c.Param("id")

	var req dto.UpdateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	admin, err := h.uc.UpdateAdminUser(usecase.UpdateAdminInput{
		RequesterRole:         entity.AdminRole(claims.Role),
		RequesterRestaurantID: claims.RestaurantID,
		TargetID:              targetID,
		Username:              req.Username,
		Password:              req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrForbidden):
			c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "forbidden", Message: "Operación no permitida."})
		case errors.Is(err, usecase.ErrUserNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "Usuario no encontrado."})
		default:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.ToAdminUserResponse(admin))
}

// ResetPassword godoc
// @Summary      Blanquear contraseña de usuario administrador
// @Description  Genera una nueva contraseña aleatoria para el usuario. Superadmin puede blanquear cualquier usuario; owner solo employees de su restaurante.
// @Tags         admin
// @Produce      json
// @Param        id   path      string  true  "ID del usuario"
// @Success      200  {object}  dto.ResetPasswordResponse
// @Failure      403  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/admin/users/{id}/reset-password [post]
func (h *AdminHandler) ResetPassword(c *gin.Context) {
	claims := mw.GetAdminClaims(c)
	targetID := c.Param("id")

	out, err := h.uc.ResetPassword(usecase.ResetPasswordInput{
		RequesterRole:         entity.AdminRole(claims.Role),
		RequesterRestaurantID: claims.RestaurantID,
		TargetID:              targetID,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrForbidden):
			c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "forbidden", Message: "Operación no permitida."})
		case errors.Is(err, usecase.ErrUserNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "Usuario no encontrado."})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.ResetPasswordResponse{NewPassword: out.NewPassword})
}
