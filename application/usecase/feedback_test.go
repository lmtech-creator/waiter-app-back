package usecase

import (
	"fmt"
	"testing"

	"github.com/waiter/back/domain/entity"
	"github.com/waiter/back/mocks"
)

func TestCreateFeedback_Success(t *testing.T) {
	tableRepo := mocks.NewTableRepo()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1", RestaurantID: "r1"}

	feedbackRepo := mocks.NewFeedbackRepo()
	uc := NewFeedbackUseCase(feedbackRepo, tableRepo)

	input := CreateFeedbackInput{TableID: "t1", Score: 4, Comment: "Great!"}
	fb, err := uc.CreateFeedback(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if fb.TableID != "t1" {
		t.Errorf("expected table_id t1, got %s", fb.TableID)
	}
	if fb.Score != 4 {
		t.Errorf("expected score 4, got %d", fb.Score)
	}
	if fb.Comment != "Great!" {
		t.Errorf("expected comment 'Great!', got %s", fb.Comment)
	}
	if fb.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestCreateFeedback_TableNotFound(t *testing.T) {
	tableRepo := mocks.NewTableRepo()
	feedbackRepo := mocks.NewFeedbackRepo()
	uc := NewFeedbackUseCase(feedbackRepo, tableRepo)

	input := CreateFeedbackInput{TableID: "nonexistent", Score: 3}
	_, err := uc.CreateFeedback(input)
	if err == nil {
		t.Fatal("expected error for nonexistent table")
	}
}

func TestCreateFeedback_InvalidScoreLow(t *testing.T) {
	tableRepo := mocks.NewTableRepo()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1"}

	feedbackRepo := mocks.NewFeedbackRepo()
	uc := NewFeedbackUseCase(feedbackRepo, tableRepo)

	input := CreateFeedbackInput{TableID: "t1", Score: 0}
	_, err := uc.CreateFeedback(input)
	if err == nil {
		t.Fatal("expected error for score < 1")
	}
}

func TestCreateFeedback_InvalidScoreHigh(t *testing.T) {
	tableRepo := mocks.NewTableRepo()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1"}

	feedbackRepo := mocks.NewFeedbackRepo()
	uc := NewFeedbackUseCase(feedbackRepo, tableRepo)

	input := CreateFeedbackInput{TableID: "t1", Score: 6}
	_, err := uc.CreateFeedback(input)
	if err == nil {
		t.Fatal("expected error for score > 5")
	}
}

func TestCreateFeedback_RepoError(t *testing.T) {
	tableRepo := mocks.NewTableRepo()
	tableRepo.Tables["t1"] = &entity.Table{ID: "t1"}

	feedbackRepo := mocks.NewFeedbackRepo()
	feedbackRepo.CreateFn = func(f *entity.Feedback) error {
		return fmt.Errorf("db error")
	}
	uc := NewFeedbackUseCase(feedbackRepo, tableRepo)

	input := CreateFeedbackInput{TableID: "t1", Score: 3}
	_, err := uc.CreateFeedback(input)
	if err == nil {
		t.Fatal("expected error from repo failure")
	}
}

func TestGetFeedbackByTable_Success(t *testing.T) {
	tableRepo := mocks.NewTableRepo()
	feedbackRepo := mocks.NewFeedbackRepo()
	feedbackRepo.Feedbacks["f1"] = &entity.Feedback{ID: "f1", TableID: "t1", Score: 5}
	feedbackRepo.Feedbacks["f2"] = &entity.Feedback{ID: "f2", TableID: "t1", Score: 3}
	feedbackRepo.Feedbacks["f3"] = &entity.Feedback{ID: "f3", TableID: "t2", Score: 4}

	uc := NewFeedbackUseCase(feedbackRepo, tableRepo)

	list, err := uc.GetFeedbackByTable("t1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 feedbacks, got %d", len(list))
	}
}

func TestGetFeedbackByTable_Empty(t *testing.T) {
	feedbackRepo := mocks.NewFeedbackRepo()
	tableRepo := mocks.NewTableRepo()
	uc := NewFeedbackUseCase(feedbackRepo, tableRepo)

	list, err := uc.GetFeedbackByTable("t99")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 feedbacks, got %d", len(list))
	}
}

func TestGetFeedbackByTable_RepoError(t *testing.T) {
	feedbackRepo := mocks.NewFeedbackRepo()
	feedbackRepo.FindByTableFn = func(tableID string) ([]entity.Feedback, error) {
		return nil, fmt.Errorf("db error")
	}
	tableRepo := mocks.NewTableRepo()
	uc := NewFeedbackUseCase(feedbackRepo, tableRepo)

	_, err := uc.GetFeedbackByTable("t1")
	if err == nil {
		t.Fatal("expected error from repo failure")
	}
}
