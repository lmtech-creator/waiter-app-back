package dto

import (
	"testing"
	"time"

	"github.com/waiter/back/domain/entity"
)

func TestToRestaurantResponse(t *testing.T) {
	r := &entity.Restaurant{ID: "r1", Name: "Pizza House", Plan: "premium"}
	resp := ToRestaurantResponse(r)

	if resp.ID != "r1" {
		t.Errorf("expected ID r1, got %s", resp.ID)
	}
	if resp.Name != "Pizza House" {
		t.Errorf("expected name 'Pizza House', got %s", resp.Name)
	}
	if resp.Plan != "premium" {
		t.Errorf("expected plan 'premium', got %s", resp.Plan)
	}
}

func TestToTableResponse(t *testing.T) {
	tbl := &entity.Table{ID: "t1", Number: 5, RestaurantID: "r1", QRCode: "qr-123"}
	resp := ToTableResponse(tbl)

	if resp.ID != "t1" || resp.Number != 5 || resp.RestaurantID != "r1" || resp.QRCode != "qr-123" {
		t.Errorf("unexpected table response: %+v", resp)
	}
}

func TestToTableListResponse(t *testing.T) {
	tables := []entity.Table{
		{ID: "t1", Number: 1, RestaurantID: "r1", QRCode: "qr1"},
		{ID: "t2", Number: 2, RestaurantID: "r1", QRCode: "qr2"},
	}
	resp := ToTableListResponse(tables)

	if len(resp) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp))
	}
	if resp[0].ID != "t1" || resp[1].ID != "t2" {
		t.Error("unexpected table IDs in list response")
	}
}

func TestToTableListResponse_Empty(t *testing.T) {
	resp := ToTableListResponse([]entity.Table{})
	if len(resp) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp))
	}
}

func TestToRequestResponse(t *testing.T) {
	now := time.Now()
	req := &entity.Request{
		ID: "req1", TableID: "t1", Type: entity.CallWaiter,
		Status: entity.Pending, CreatedAt: now,
	}
	resp := ToRequestResponse(req)

	if resp.ID != "req1" || resp.TableID != "t1" {
		t.Errorf("unexpected request response: %+v", resp)
	}
	if resp.Type != entity.CallWaiter || resp.Status != entity.Pending {
		t.Error("type/status mismatch")
	}
	if !resp.CreatedAt.Equal(now) {
		t.Error("created_at mismatch")
	}
}

func TestToRequestListResponse(t *testing.T) {
	requests := []entity.Request{
		{ID: "r1", TableID: "t1", Status: entity.Pending},
		{ID: "r2", TableID: "t1", Status: entity.Done},
	}
	resp := ToRequestListResponse(requests)

	if len(resp) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp))
	}
}

func TestToRequestListResponse_Empty(t *testing.T) {
	resp := ToRequestListResponse([]entity.Request{})
	if len(resp) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp))
	}
}

func TestToFeedbackResponse(t *testing.T) {
	now := time.Now()
	fb := &entity.Feedback{ID: "f1", TableID: "t1", Score: 4, Comment: "Good", CreatedAt: now}
	resp := ToFeedbackResponse(fb)

	if resp.ID != "f1" || resp.Score != 4 || resp.Comment != "Good" {
		t.Errorf("unexpected feedback response: %+v", resp)
	}
	if !resp.CreatedAt.Equal(now) {
		t.Error("created_at mismatch")
	}
}

func TestToFeedbackResponse_EmptyComment(t *testing.T) {
	fb := &entity.Feedback{ID: "f1", TableID: "t1", Score: 3}
	resp := ToFeedbackResponse(fb)

	if resp.Comment != "" {
		t.Errorf("expected empty comment, got '%s'", resp.Comment)
	}
}

func TestToFeedbackListResponse(t *testing.T) {
	feedbacks := []entity.Feedback{
		{ID: "f1", TableID: "t1", Score: 5},
		{ID: "f2", TableID: "t1", Score: 3},
	}
	resp := ToFeedbackListResponse(feedbacks)

	if len(resp) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp))
	}
}

func TestToFeedbackListResponse_Empty(t *testing.T) {
	resp := ToFeedbackListResponse([]entity.Feedback{})
	if len(resp) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp))
	}
}
