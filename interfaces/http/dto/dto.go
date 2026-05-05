package dto

import (
	"time"

	"github.com/waiter/back/domain/entity"
)

// ─── Request DTOs ───

type SessionRequest struct {
	QRCode string `json:"qr_code"`
}

type SessionResponse struct {
	SessionToken string      `json:"session_token"`
	Table        TablePublic `json:"table"`
}

type TablePublic struct {
	Number int `json:"number"`
}

type CreateRestaurantRequest struct {
	Name string `json:"name" binding:"required"`
	Plan string `json:"plan"`
}

type CreateTableRequest struct {
	Number int `json:"number" binding:"required"`
}

type CreateRequestRequest struct {
	Type entity.RequestType `json:"type" binding:"required"`
}

type UpdateRequestStatusRequest struct {
	Status entity.RequestStatus `json:"status" binding:"required"`
}

type CreateFeedbackRequest struct {
	TableID string `json:"table_id" binding:"required"`
	Score   int    `json:"score" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

// ─── Response DTOs ───

type RestaurantResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Plan string `json:"plan"`
}

type TableResponse struct {
	ID           string `json:"id"`
	Number       int    `json:"number"`
	RestaurantID string `json:"restaurant_id"`
	QRCode       string `json:"qr_code"`
}

type RequestResponse struct {
	ID        string               `json:"id"`
	TableID   string               `json:"table_id"`
	Type      entity.RequestType   `json:"type"`
	Status    entity.RequestStatus `json:"status"`
	CreatedAt time.Time            `json:"created_at"`
}

type FeedbackResponse struct {
	ID        string    `json:"id"`
	TableID   string    `json:"table_id"`
	Score     int       `json:"score"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ─── Admin DTOs ───

type AdminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AdminLoginResponse struct {
	Token string `json:"token"`
}

type CreateAdminUserRequest struct {
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	Role         string `json:"role" binding:"required"`
	RestaurantID string `json:"restaurant_id"`
}

type AdminUserResponse struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Role         string    `json:"role"`
	RestaurantID *string   `json:"restaurant_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// ─── Mappers: Entity → Response ───

func ToRestaurantResponse(r *entity.Restaurant) RestaurantResponse {
	return RestaurantResponse{
		ID:   r.ID,
		Name: r.Name,
		Plan: r.Plan,
	}
}

func ToTableResponse(t *entity.Table) TableResponse {
	return TableResponse{
		ID:           t.ID,
		Number:       t.Number,
		RestaurantID: t.RestaurantID,
		QRCode:       t.QRCode,
	}
}

func ToTableListResponse(tables []entity.Table) []TableResponse {
	out := make([]TableResponse, len(tables))
	for i, t := range tables {
		out[i] = ToTableResponse(&t)
	}
	return out
}

func ToRequestResponse(r *entity.Request) RequestResponse {
	return RequestResponse{
		ID:        r.ID,
		TableID:   r.TableID,
		Type:      r.Type,
		Status:    r.Status,
		CreatedAt: r.CreatedAt,
	}
}

func ToRequestListResponse(requests []entity.Request) []RequestResponse {
	out := make([]RequestResponse, len(requests))
	for i, r := range requests {
		out[i] = ToRequestResponse(&r)
	}
	return out
}

func ToFeedbackResponse(f *entity.Feedback) FeedbackResponse {
	return FeedbackResponse{
		ID:        f.ID,
		TableID:   f.TableID,
		Score:     f.Score,
		Comment:   f.Comment,
		CreatedAt: f.CreatedAt,
	}
}

func ToFeedbackListResponse(feedbacks []entity.Feedback) []FeedbackResponse {
	out := make([]FeedbackResponse, len(feedbacks))
	for i, f := range feedbacks {
		out[i] = ToFeedbackResponse(&f)
	}
	return out
}

func ToAdminUserResponse(a *entity.AdminUser) AdminUserResponse {
	return AdminUserResponse{
		ID:           a.ID,
		Username:     a.Username,
		Role:         string(a.Role),
		RestaurantID: a.RestaurantID,
		CreatedAt:    a.CreatedAt,
	}
}

func ToAdminUserListResponse(users []entity.AdminUser) []AdminUserResponse {
	out := make([]AdminUserResponse, len(users))
	for i, u := range users {
		out[i] = ToAdminUserResponse(&u)
	}
	return out
}
