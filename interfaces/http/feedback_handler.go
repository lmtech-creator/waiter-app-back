package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/application/usecase"
	"github.com/waiter/back/interfaces/http/dto"
)

type FeedbackHandler struct {
	uc *usecase.FeedbackUseCase
}

func NewFeedbackHandler(uc *usecase.FeedbackUseCase) *FeedbackHandler {
	return &FeedbackHandler{uc: uc}
}

// Create godoc
// @Summary      Registrar feedback
// @Description  Registra el feedback de un cliente sobre el servicio
// @Tags         feedback
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateFeedbackRequest  true  "Datos del feedback"
// @Success      201      {object}  dto.FeedbackResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      422      {object}  dto.ErrorResponse
// @Router       /api/v1/feedback [post]
func (h *FeedbackHandler) Create(c *gin.Context) {
	var req dto.CreateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: err.Error()})
		return
	}

	fb, err := h.uc.CreateFeedback(usecase.CreateFeedbackInput{
		TableID:   req.TableID,
		Score:     req.Score,
		Comment:   req.Comment,
		RequestID: req.RequestID,
	})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: "unprocessable_entity", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ToFeedbackResponse(fb))
}

// GetByRestaurant godoc
// @Summary      Listar feedbacks del restaurante
// @Description  Retorna todos los feedbacks del restaurante (solo superadmin y owner)
// @Tags         feedback
// @Produce      json
// @Param        restaurantId  path  string  true  "ID del restaurante"
// @Success      200  {array}   dto.FeedbackResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/restaurants/{restaurantId}/feedback [get]
func (h *FeedbackHandler) GetByRestaurant(c *gin.Context) {
	restaurantID := c.Param("restaurantId")

	list, err := h.uc.GetFeedbackByRestaurant(restaurantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToFeedbackListResponse(list))
}
