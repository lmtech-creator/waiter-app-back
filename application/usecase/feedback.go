package usecase

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/domain/repository"
)

type FeedbackUseCase struct {
	feedbackRepo repository.FeedbackRepository
	tableRepo    repository.TableRepository
}

func NewFeedbackUseCase(fr repository.FeedbackRepository, tr repository.TableRepository) *FeedbackUseCase {
	return &FeedbackUseCase{feedbackRepo: fr, tableRepo: tr}
}

type CreateFeedbackInput struct {
	TableID   string  `json:"table_id" binding:"required"`
	Score     int     `json:"score" binding:"required,min=1,max=5"`
	Comment   string  `json:"comment"`
	RequestID *string `json:"request_id,omitempty"`
}

func (uc *FeedbackUseCase) CreateFeedback(input CreateFeedbackInput) (*entity.Feedback, error) {
	if _, err := uc.tableRepo.FindByID(input.TableID); err != nil {
		return nil, fmt.Errorf("table not found: %w", err)
	}

	if input.Score < 1 || input.Score > 5 {
		return nil, fmt.Errorf("score must be between 1 and 5")
	}

	fb := &entity.Feedback{
		ID:        uuid.New().String(),
		TableID:   input.TableID,
		RequestID: input.RequestID,
		Score:     input.Score,
		Comment:   input.Comment,
	}

	if err := uc.feedbackRepo.Create(fb); err != nil {
		return nil, fmt.Errorf("failed to create feedback: %w", err)
	}

	return fb, nil
}

func (uc *FeedbackUseCase) GetFeedbackByTable(tableID string) ([]entity.Feedback, error) {
	return uc.feedbackRepo.FindByTableID(tableID)
}

func (uc *FeedbackUseCase) GetFeedbackByRestaurant(restaurantID string) ([]entity.Feedback, error) {
	return uc.feedbackRepo.FindByRestaurantID(restaurantID)
}
