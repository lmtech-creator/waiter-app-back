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
		TableID: req.TableID,
		Score:   req.Score,
		Comment: req.Comment,
	})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: "unprocessable_entity", Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ToFeedbackResponse(fb))
}
