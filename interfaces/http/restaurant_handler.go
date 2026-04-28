package http

import (
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
// @Router       /api/v1/restaurants [post]
func (h *RestaurantHandler) Create(c *gin.Context) {
	var req dto.CreateRestaurantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	r, err := h.uc.CreateRestaurant(usecase.CreateRestaurantInput{
		Name: req.Name,
		Plan: req.Plan,
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
// @Failure      422           {object}  dto.ErrorResponse
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
