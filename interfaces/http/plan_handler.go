package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/waiter/back/application/usecase"
	"github.com/waiter/back/interfaces/http/dto"
)

type PlanHandler struct {
	uc *usecase.PlanUseCase
}

func NewPlanHandler(uc *usecase.PlanUseCase) *PlanHandler {
	return &PlanHandler{uc: uc}
}

// Create godoc
// @Summary      Crear plan
// @Description  Crea un nuevo plan configurable (solo superadmin)
// @Tags         restaurants
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreatePlanRequest  true  "Datos del plan"
// @Success      201      {object}  dto.PlanResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      422      {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/plans [post]
func (h *PlanHandler) Create(c *gin.Context) {
	var req dto.CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	plan, err := h.uc.CreatePlan(usecase.CreatePlanInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToPlanResponse(plan))
}

// GetAll godoc
// @Summary      Listar planes
// @Description  Retorna todos los planes configurables (solo superadmin)
// @Tags         restaurants
// @Produce      json
// @Success      200  {array}  dto.PlanResponse
// @Failure      500  {object} dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/plans [get]
func (h *PlanHandler) GetAll(c *gin.Context) {
	plans, err := h.uc.GetAllPlans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ToPlanListResponse(plans))
}

// Get godoc
// @Summary      Obtener plan
// @Description  Retorna un plan por su ID (solo superadmin)
// @Tags         restaurants
// @Produce      json
// @Param        planId  path  string  true  "ID del plan"
// @Success      200  {object}  dto.PlanResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/plans/{planId} [get]
func (h *PlanHandler) Get(c *gin.Context) {
	plan, err := h.uc.GetPlan(c.Param("planId"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToPlanResponse(plan))
}

// Update godoc
// @Summary      Actualizar plan
// @Description  Actualiza un plan configurable (solo superadmin)
// @Tags         restaurants
// @Accept       json
// @Produce      json
// @Param        planId   path      string                true  "ID del plan"
// @Param        request  body      dto.UpdatePlanRequest true  "Datos a actualizar"
// @Success      200      {object}  dto.PlanResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      404      {object}  dto.ErrorResponse
// @Failure      422      {object}  dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/plans/{planId} [patch]
func (h *PlanHandler) Update(c *gin.Context) {
	var req dto.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	plan, err := h.uc.UpdatePlan(c.Param("planId"), usecase.UpdatePlanInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ToPlanResponse(plan))
}

// Delete godoc
// @Summary      Eliminar plan
// @Description  Elimina un plan configurable si no está en uso (solo superadmin)
// @Tags         restaurants
// @Produce      json
// @Param        planId  path  string  true  "ID del plan"
// @Success      204  "Plan eliminado"
// @Failure      404  {object} dto.ErrorResponse
// @Failure      409  {object} dto.ErrorResponse
// @Security     AdminToken
// @Router       /api/v1/plans/{planId} [delete]
func (h *PlanHandler) Delete(c *gin.Context) {
	if err := h.uc.DeletePlan(c.Param("planId")); err != nil {
		h.writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *PlanHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrPlanNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "not_found", Message: err.Error()})
	case errors.Is(err, usecase.ErrPlanNameExists), errors.Is(err, usecase.ErrPlanInUse):
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "conflict", Message: err.Error()})
	case errors.Is(err, usecase.ErrPlanNameRequired):
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: "unprocessable_entity", Message: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: err.Error()})
	}
}
