package usecase

import (
	"fmt"
	"testing"

	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/mocks"
)

func TestCreateRequest_Success(t *testing.T) {
	tableRepo := mocks.NewTableRepo()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", RestaurantID: "r1"}

	requestRepo := mocks.NewRequestRepo()
	notifier := mocks.NewNotifier()
	uc := NewRequestUseCase(requestRepo, tableRepo, notifier)

	input := CreateRequestInput{TableID: "t1", Type: entity.CallWaiter}
	req, err := uc.CreateRequest(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if req.TableID != "t1" {
		t.Errorf("expected table_id t1, got %s", req.TableID)
	}
	if req.Type != entity.CallWaiter {
		t.Errorf("expected type CALL_WAITER, got %s", req.Type)
	}
	if req.Status != entity.Pending {
		t.Errorf("expected status PENDING, got %s", req.Status)
	}
	if req.ID == "" {
		t.Error("expected non-empty ID")
	}
	if len(notifier.Events) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifier.Events))
	}
	if notifier.Events[0].RestaurantID != "r1" {
		t.Errorf("expected notification for r1, got %s", notifier.Events[0].RestaurantID)
	}
}

func TestCreateRequest_AllTypes(t *testing.T) {
	types := []entity.RequestType{entity.CallWaiter, entity.AskBill, entity.AskHelp}
	for _, rt := range types {
		t.Run(string(rt), func(t *testing.T) {
			tableRepo := mocks.NewTableRepo()
			tableRepo.Tables["t1"] = &entity.Table{ID: "t1", RestaurantID: "r1"}
			requestRepo := mocks.NewRequestRepo()
			notifier := mocks.NewNotifier()
			uc := NewRequestUseCase(requestRepo, tableRepo, notifier)

			req, err := uc.CreateRequest(CreateRequestInput{TableID: "t1", Type: rt})
			if err != nil {
				t.Fatalf("expected no error for type %s, got %v", rt, err)
			}
			if req.Type != rt {
				t.Errorf("expected type %s, got %s", rt, req.Type)
			}
		})
	}
}

func TestCreateRequest_InvalidType(t *testing.T) {
	tableRepo := mocks.NewTableRepo()
	requestRepo := mocks.NewRequestRepo()
	notifier := mocks.NewNotifier()
	uc := NewRequestUseCase(requestRepo, tableRepo, notifier)

	_, err := uc.CreateRequest(CreateRequestInput{TableID: "t1", Type: "INVALID"})
	if err == nil {
		t.Fatal("expected error for invalid request type")
	}
}

func TestCreateRequest_TableNotFound(t *testing.T) {
	tableRepo := mocks.NewTableRepo()
	requestRepo := mocks.NewRequestRepo()
	notifier := mocks.NewNotifier()
	uc := NewRequestUseCase(requestRepo, tableRepo, notifier)

	_, err := uc.CreateRequest(CreateRequestInput{TableID: "nonexistent", Type: entity.CallWaiter})
	if err == nil {
		t.Fatal("expected error for nonexistent table")
	}
}

func TestCreateRequest_RepoError(t *testing.T) {
	tableRepo := mocks.NewTableRepo()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", RestaurantID: "r1"}

	requestRepo := mocks.NewRequestRepo()
	requestRepo.CreateFn = func(r *entity.Request) error {
		return fmt.Errorf("db error")
	}
	notifier := mocks.NewNotifier()
	uc := NewRequestUseCase(requestRepo, tableRepo, notifier)

	_, err := uc.CreateRequest(CreateRequestInput{TableID: "t1", Type: entity.CallWaiter})
	if err == nil {
		t.Fatal("expected error from repo failure")
	}
}

func TestGetActiveRequests_Success(t *testing.T) {
	requestRepo := mocks.NewRequestRepo()
	requestRepo.Requests["r1"] = &entity.Request{ID: "r1", TableID: "t1", Status: entity.Pending}
	requestRepo.Requests["r2"] = &entity.Request{ID: "r2", TableID: "t1", Status: entity.Done}

	tableRepo := mocks.NewTableRepo()
	notifier := mocks.NewNotifier()
	uc := NewRequestUseCase(requestRepo, tableRepo, notifier)

	list, err := uc.GetActiveRequests("r1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 active request, got %d", len(list))
	}
}

func TestGetActiveRequests_RepoError(t *testing.T) {
	requestRepo := mocks.NewRequestRepo()
	requestRepo.FindActiveFn = func(restaurantID string) ([]entity.Request, error) {
		return nil, fmt.Errorf("db error")
	}
	tableRepo := mocks.NewTableRepo()
	notifier := mocks.NewNotifier()
	uc := NewRequestUseCase(requestRepo, tableRepo, notifier)

	_, err := uc.GetActiveRequests("r1")
	if err == nil {
		t.Fatal("expected error from repo failure")
	}
}

func TestCompleteRequest_Success(t *testing.T) {
	requestRepo := mocks.NewRequestRepo()
	requestRepo.Requests["req1"] = &entity.Request{ID: "req1", TableID: "t1", Status: entity.Pending}

	tableRepo := mocks.NewTableRepo()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", RestaurantID: "r1"}

	notifier := mocks.NewNotifier()
	uc := NewRequestUseCase(requestRepo, tableRepo, notifier)

	err := uc.CompleteRequest("req1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if requestRepo.Requests["req1"].Status != entity.Done {
		t.Error("expected request status to be DONE")
	}
	if len(notifier.Events) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifier.Events))
	}
}

func TestCompleteRequest_NotFound(t *testing.T) {
	requestRepo := mocks.NewRequestRepo()
	tableRepo := mocks.NewTableRepo()
	notifier := mocks.NewNotifier()
	uc := NewRequestUseCase(requestRepo, tableRepo, notifier)

	err := uc.CompleteRequest("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent request")
	}
}

func TestCompleteRequest_AlreadyDone(t *testing.T) {
	requestRepo := mocks.NewRequestRepo()
	requestRepo.Requests["req1"] = &entity.Request{ID: "req1", TableID: "t1", Status: entity.Done}

	tableRepo := mocks.NewTableRepo()
	notifier := mocks.NewNotifier()
	uc := NewRequestUseCase(requestRepo, tableRepo, notifier)

	err := uc.CompleteRequest("req1")
	if err == nil {
		t.Fatal("expected error for already completed request")
	}
}

func TestCompleteRequest_UpdateStatusError(t *testing.T) {
	requestRepo := mocks.NewRequestRepo()
	requestRepo.Requests["req1"] = &entity.Request{ID: "req1", TableID: "t1", Status: entity.Pending}
	requestRepo.UpdateStatusFn = func(id string, status entity.RequestStatus) error {
		return fmt.Errorf("db error")
	}

	tableRepo := mocks.NewTableRepo()
	notifier := mocks.NewNotifier()
	uc := NewRequestUseCase(requestRepo, tableRepo, notifier)

	err := uc.CompleteRequest("req1")
	if err == nil {
		t.Fatal("expected error from update status failure")
	}
}

func TestCompleteRequest_TableNotFound_StillCompletes(t *testing.T) {
	requestRepo := mocks.NewRequestRepo()
	requestRepo.Requests["req1"] = &entity.Request{ID: "req1", TableID: "t_missing", Status: entity.Pending}

	tableRepo := mocks.NewTableRepo() // table not in repo
	notifier := mocks.NewNotifier()
	uc := NewRequestUseCase(requestRepo, tableRepo, notifier)

	err := uc.CompleteRequest("req1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Notification should NOT be sent since table was not found
	if len(notifier.Events) != 0 {
		t.Errorf("expected 0 notifications when table not found, got %d", len(notifier.Events))
	}
}

func TestGetTableStatus_Success(t *testing.T) {
	requestRepo := mocks.NewRequestRepo()
	requestRepo.Requests["r1"] = &entity.Request{ID: "r1", TableID: "t1", Status: entity.Pending}
	requestRepo.Requests["r2"] = &entity.Request{ID: "r2", TableID: "t2", Status: entity.Pending}

	tableRepo := mocks.NewTableRepo()
	notifier := mocks.NewNotifier()
	uc := NewRequestUseCase(requestRepo, tableRepo, notifier)

	list, err := uc.GetTableStatus("t1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 request for table t1, got %d", len(list))
	}
}

func TestGetTableStatus_RepoError(t *testing.T) {
	requestRepo := mocks.NewRequestRepo()
	requestRepo.FindByTableFn = func(tableID string) ([]entity.Request, error) {
		return nil, fmt.Errorf("db error")
	}
	tableRepo := mocks.NewTableRepo()
	notifier := mocks.NewNotifier()
	uc := NewRequestUseCase(requestRepo, tableRepo, notifier)

	_, err := uc.GetTableStatus("t1")
	if err == nil {
		t.Fatal("expected error from repo failure")
	}
}
