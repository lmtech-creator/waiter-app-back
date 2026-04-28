package usecase

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/domain/repository"
)

type RequestUseCase struct {
	requestRepo repository.RequestRepository
	tableRepo   repository.TableRepository
	notifier    EventNotifier
}

// EventNotifier sends real-time events to connected clients.
type EventNotifier interface {
	Notify(restaurantID string, event any)
}

func NewRequestUseCase(rr repository.RequestRepository, tr repository.TableRepository, n EventNotifier) *RequestUseCase {
	return &RequestUseCase{requestRepo: rr, tableRepo: tr, notifier: n}
}

// CooldownSeconds is the minimum wait between requests from the same table.
const CooldownSeconds = 15

// CooldownError is returned when a request is rejected due to the per-table cooldown.
type CooldownError struct {
	SecondsRemaining int
}

func (e *CooldownError) Error() string {
	return fmt.Sprintf("cooldown: wait %d seconds", e.SecondsRemaining)
}

type CreateRequestInput struct {
	TableID      string             `json:"table_id"`
	RestaurantID string             `json:"restaurant_id,omitempty"`
	Type         entity.RequestType `json:"type" binding:"required"`
}

func (uc *RequestUseCase) CreateRequest(input CreateRequestInput) (*entity.Request, error) {
	if input.Type != entity.CallWaiter && input.Type != entity.AskBill && input.Type != entity.AskHelp {
		return nil, fmt.Errorf("invalid request type: %s", input.Type)
	}

	// Resolve RestaurantID: prefer from input (JWT claims), fall back to table lookup.
	restaurantID := input.RestaurantID
	if restaurantID == "" {
		table, err := uc.tableRepo.FindByID(input.TableID)
		if err != nil {
			return nil, fmt.Errorf("table not found: %w", err)
		}
		restaurantID = table.RestaurantID
	}

	// Per-table cooldown check (§5): query DB for the last request timestamp.
	last, err := uc.requestRepo.FindLastCreatedByTableID(input.TableID)
	if err == nil {
		elapsed := time.Since(last.CreatedAt)
		if elapsed < CooldownSeconds*time.Second {
			remaining := CooldownSeconds - int(elapsed.Seconds())
			return nil, &CooldownError{SecondsRemaining: remaining}
		}
	}
	// err != nil means no previous request — cooldown not applicable.

	req := &entity.Request{
		ID:      uuid.New().String(),
		TableID: input.TableID,
		Type:    input.Type,
		Status:  entity.Pending,
	}

	if err := uc.requestRepo.Create(req); err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	uc.notifier.Notify(restaurantID, map[string]any{
		"event":   "new_request",
		"request": req,
	})

	return req, nil
}

func (uc *RequestUseCase) GetActiveRequests(restaurantID string) ([]entity.Request, error) {
	return uc.requestRepo.FindActiveByRestaurantID(restaurantID)
}

func (uc *RequestUseCase) CompleteRequest(id string) error {
	req, err := uc.requestRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("request not found: %w", err)
	}

	if req.Status == entity.Done {
		return fmt.Errorf("request already completed")
	}

	if err := uc.requestRepo.UpdateStatus(id, entity.Done); err != nil {
		return fmt.Errorf("failed to update request: %w", err)
	}

	table, err := uc.tableRepo.FindByID(req.TableID)
	if err == nil {
		uc.notifier.Notify(table.RestaurantID, map[string]any{
			"event":      "request_completed",
			"request_id": id,
		})
	}

	return nil
}

func (uc *RequestUseCase) GetTableStatus(tableID string) ([]entity.Request, error) {
	return uc.requestRepo.FindByTableID(tableID)
}
