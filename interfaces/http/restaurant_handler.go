package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/application/usecase"
	"github.com/waiter/back/interfaces/http/dto"
	"go.uber.org/zap"
)

type RestaurantHandler struct {
	uc *usecase.RestaurantUseCase
}

func NewRestaurantHandler(uc *usecase.RestaurantUseCase) *RestaurantHandler {
	return &RestaurantHandler{uc: uc}
}

// Create godoc
// @Summary      Crear restaurante
// @Description  Crea un nuevo restaurante
// @Tags         restaurants
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateRestaurantRequest  true  "Datos del restaurante"
// @Success      201      {object}  dto.RestaurantResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      422      {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/restaurants [post]
func (h *RestaurantHandler) Create(c *gin.Context) {
	var req dto.CreateRestaurantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	r, err := h.uc.CreateRestaurant(usecase.CreateRestaurantInput{
		Name:          req.Name,
		Plan:          req.Plan,
		QrBannerText:  req.QrBannerText,
		QrFooterItems: req.QrFooterItems,
	})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: "unprocessable_entity", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ToRestaurantResponse(r))
}

// Get godoc
// @Summary      Obtener restaurante
// @Description  Retorna un restaurante por su ID
// @Tags         restaurants
// @Produce      json
// @Param        restaurantId  path      string  true  "ID del restaurante"
// @Success      200           {object}  dto.RestaurantResponse
// @Failure      404           {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/restaurants/{restaurantId} [get]
func (h *RestaurantHandler) Get(c *gin.Context) {
	id := c.Param("restaurantId")

	r, err := h.uc.GetRestaurant(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "restaurant not found"})
		return
	}

	c.JSON(http.StatusOK, dto.ToRestaurantResponse(r))
}

// GetAll godoc
// @Summary      Listar todos los restaurantes
// @Description  Retorna todos los restaurantes del sistema (solo superadmin)
// @Tags         restaurants
// @Produce      json
// @Success      200  {array}   dto.RestaurantResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/restaurants [get]
func (h *RestaurantHandler) GetAll(c *gin.Context) {
	restaurants, err := h.uc.GetAllRestaurants()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	out := make([]dto.RestaurantResponse, len(restaurants))
	for i, r := range restaurants {
		out[i] = dto.ToRestaurantResponse(&r)
	}
	c.JSON(http.StatusOK, out)
}

// CreateTable godoc
// @Summary      Crear mesa
// @Description  Crea una nueva mesa para un restaurante
// @Tags         tables
// @Accept       json
// @Produce      json
// @Param        restaurantId  path      string                  true  "ID del restaurante"
// @Param        request       body      dto.CreateTableRequest  true  "Datos de la mesa"
// @Success      201           {object}  dto.TableResponse
// @Failure      400           {object}  dto.ErrorResponse
// @Failure      409           {object}  dto.ErrorResponse
// @Failure      422           {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/restaurants/{restaurantId}/tables [post]
func (h *RestaurantHandler) CreateTable(c *gin.Context) {
	restaurantID := c.Param("restaurantId")

	var req dto.CreateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	t, err := h.uc.CreateTable(usecase.CreateTableInput{
		Number:       req.Number,
		RestaurantID: restaurantID,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrTableNumberExists) {
			c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "conflict", Message: "ya existe una mesa con ese número en el restaurante"})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: "unprocessable_entity", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ToTableResponse(t))
}

// GetTables godoc
// @Summary      Listar mesas
// @Description  Retorna las mesas de un restaurante
// @Tags         tables
// @Produce      json
// @Param        restaurantId  path      string  true  "ID del restaurante"
// @Success      200           {array}   dto.TableResponse
// @Failure      500           {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/restaurants/{restaurantId}/tables [get]
func (h *RestaurantHandler) GetTables(c *gin.Context) {
	restaurantID := c.Param("restaurantId")

	tables, err := h.uc.GetTables(restaurantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: "failed to retrieve tables"})
		return
	}

	c.JSON(http.StatusOK, dto.ToTableListResponse(tables))
}

// RegenerateQR godoc
// @Summary      Regenerar QR de mesa
// @Description  Genera un nuevo código QR aleatorio para una mesa (admin)
// @Tags         tables
// @Produce      json
// @Param        id  path      string  true  "ID de la mesa"
// @Success      200  {object}  map[string]string
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/admin/tables/{id}/regenerate-qr [post]
func (h *RestaurantHandler) RegenerateQR(c *gin.Context) {
	tableID := c.Param("id")

	newCode, err := h.uc.RegenerateQR(tableID)
	if err != nil {
		if strings.Contains(err.Error(), "table not found") {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "Mesa no encontrada."})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: "No se pudo regenerar el QR."})
		return
	}

	zap.L().Info("qr regenerated", zap.String("table_id", tableID))
	c.JSON(http.StatusOK, gin.H{"qr_code": newCode})
}

// InactivateTable godoc
// @Summary      Inactivar mesa
// @Description  Marca una mesa como inactiva para bloquear nuevas sesiones QR
// @Tags         tables
// @Produce      json
// @Param        restaurantId  path      string  true  "ID del restaurante"
// @Param        id            path      string  true  "ID de la mesa"
// @Success      200           {object}  map[string]any
// @Failure      403           {object}  dto.ErrorResponse
// @Failure      404           {object}  dto.ErrorResponse
// @Failure      500           {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/restaurants/{restaurantId}/tables/{id}/inactivate [post]
func (h *RestaurantHandler) InactivateTable(c *gin.Context) {
	restaurantID := c.Param("restaurantId")
	tableID := c.Param("id")

	err := h.uc.InactivateTable(restaurantID, tableID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrTableNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "Mesa no encontrada."})
		case errors.Is(err, usecase.ErrTableRestaurantMismatch):
			c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "forbidden", Message: "La mesa no pertenece al restaurante."})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: "No se pudo inactivar la mesa."})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"table_id": tableID, "is_active": false})
}

// ActivateTable godoc
// @Summary      Activar mesa
// @Description  Marca una mesa como activa para habilitar nuevas sesiones QR
// @Tags         tables
// @Produce      json
// @Param        restaurantId  path      string  true  "ID del restaurante"
// @Param        id            path      string  true  "ID de la mesa"
// @Success      200           {object}  map[string]any
// @Failure      403           {object}  dto.ErrorResponse
// @Failure      404           {object}  dto.ErrorResponse
// @Failure      500           {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/restaurants/{restaurantId}/tables/{id}/activate [post]
func (h *RestaurantHandler) ActivateTable(c *gin.Context) {
	restaurantID := c.Param("restaurantId")
	tableID := c.Param("id")

	err := h.uc.ActivateTable(restaurantID, tableID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrTableNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: "Mesa no encontrada."})
		case errors.Is(err, usecase.ErrTableRestaurantMismatch):
			c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "forbidden", Message: "La mesa no pertenece al restaurante."})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: "No se pudo activar la mesa."})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"table_id": tableID, "is_active": true})
}
