package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/application/usecase"
	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/interfaces/http/dto"
)

type RequestHandler struct {
	uc *usecase.RequestUseCase
}

func NewRequestHandler(uc *usecase.RequestUseCase) *RequestHandler {
	return &RequestHandler{uc: uc}
}

// Create godoc
// @Summary      Crear solicitud
// @Description  Crea una nueva solicitud de mesa (llamar mozo, pedir cuenta, ayuda)
// @Tags         requests
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateRequestRequest  true  "Datos de la solicitud"
// @Success      201      {object}  dto.RequestResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      422      {object}  dto.ErrorResponse
// @Router       /api/v1/requests [post]
func (h *RequestHandler) Create(c *gin.Context) {
	var req dto.CreateRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	result, err := h.uc.CreateRequest(usecase.CreateRequestInput{
		TableID: req.TableID,
		Type:    req.Type,
	})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: "unprocessable_entity", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ToRequestResponse(result))
}

// GetActive godoc
// @Summary      Obtener solicitudes activas
// @Description  Retorna las solicitudes activas de un restaurante
// @Tags         requests
// @Produce      json
// @Param        restaurantId  path      string  true  "ID del restaurante"
// @Success      200           {array}   dto.RequestResponse
// @Failure      500           {object}  dto.ErrorResponse
// @Router       /api/v1/restaurants/{restaurantId}/requests/active [get]
func (h *RequestHandler) GetActive(c *gin.Context) {
	restaurantID := c.Param("restaurantId")

	requests, err := h.uc.GetActiveRequests(restaurantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: "failed to retrieve requests"})
		return
	}

	c.JSON(http.StatusOK, dto.ToRequestListResponse(requests))
}

// Complete godoc
// @Summary      Actualizar estado de solicitud
// @Description  Actualiza el estado de una solicitud (marcar como atendida)
// @Tags         requests
// @Accept       json
// @Produce      json
// @Param        requestId  path      string                        true  "ID de la solicitud"
// @Param        request    body      dto.UpdateRequestStatusRequest  true  "Nuevo estado"
// @Success      204
// @Failure      400        {object}  dto.ErrorResponse
// @Failure      422        {object}  dto.ErrorResponse
// @Router       /api/v1/requests/{requestId} [patch]
func (h *RequestHandler) Complete(c *gin.Context) {
	id := c.Param("requestId")

	var req dto.UpdateRequestStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	if req.Status != entity.Done {
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: "invalid_status", Message: "only DONE status transition is supported"})
		return
	}

	if err := h.uc.CompleteRequest(id); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: "unprocessable_entity", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetTableStatus godoc
// @Summary      Obtener estado de mesa
// @Description  Retorna las solicitudes de una mesa específica
// @Tags         requests
// @Produce      json
// @Param        tableId  path      string  true  "ID de la mesa"
// @Success      200      {array}   dto.RequestResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/v1/tables/{tableId}/status [get]
func (h *RequestHandler) GetTableStatus(c *gin.Context) {
	tableID := c.Param("tableId")

	requests, err := h.uc.GetTableStatus(tableID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: "failed to retrieve table status"})
		return
	}

	c.JSON(http.StatusOK, dto.ToRequestListResponse(requests))
}
