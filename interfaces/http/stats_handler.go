package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/application/usecase"
	"github.com/waiter/back/interfaces/http/dto"
)

type StatsHandler struct {
	uc *usecase.StatsUseCase
}

func NewStatsHandler(uc *usecase.StatsUseCase) *StatsHandler {
	return &StatsHandler{uc: uc}
}

// Get godoc
// @Summary      Estadísticas del restaurante
// @Description  Retorna estadísticas operativas del restaurante en un rango de fechas
// @Tags         restaurants
// @Produce      json
// @Param        restaurantId  path    string  true   "ID del restaurante"
// @Param        since         query   string  true   "Fecha inicio (RFC3339)"
// @Param        until         query   string  true   "Fecha fin (RFC3339)"
// @Success      200  {object}  dto.StatsResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/restaurants/{restaurantId}/stats [get]
func (h *StatsHandler) Get(c *gin.Context) {
	restaurantID := c.Param("restaurantId")

	sinceStr := c.Query("since")
	untilStr := c.Query("until")

	since, err := time.Parse(time.RFC3339, sinceStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: "since debe ser RFC3339 (ej: 2026-01-01T00:00:00Z)"})
		return
	}

	until, err := time.Parse(time.RFC3339, untilStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: "until debe ser RFC3339 (ej: 2026-01-01T00:00:00Z)"})
		return
	}

	stats, err := h.uc.GetStats(restaurantID, since, until)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToStatsResponse(stats))
}
