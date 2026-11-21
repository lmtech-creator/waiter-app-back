package dto

import (
	"time"

	"github.com/waiter/back/application/usecase"
	"github.com/waiter/back/domain/entity"
)

// ─── Request DTOs ───

type SessionRequest struct {
	QRCode string `json:"qr_code"`
}

type SessionResponse struct {
	SessionToken string           `json:"session_token"`
	Table        TablePublic      `json:"table"`
	Restaurant   RestaurantPublic `json:"restaurant"`
}

type TablePublic struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
}

type RestaurantPublic struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	MenuURL string `json:"menu_url"`
}

type CreateRestaurantRequest struct {
	Name          string   `json:"name" binding:"required"`
	PlanID        string   `json:"plan_id"`
	Plan          string   `json:"plan"`
	MenuURL       string   `json:"menu_url"`
	QrBannerText  string   `json:"qr_banner_text"`
	QrFooterItems []string `json:"qr_footer_items"`
}

type UpdateRestaurantRequest struct {
	MenuURL *string `json:"menu_url"`
}

type CreatePlanRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdatePlanRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
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
	TableID   string  `json:"table_id" binding:"required"`
	Score     int     `json:"score" binding:"required,min=1,max=5"`
	Comment   string  `json:"comment"`
	RequestID *string `json:"request_id,omitempty"`
}

// ─── Response DTOs ───

type RestaurantResponse struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	PlanID        string   `json:"plan_id"`
	Plan          string   `json:"plan"`
	MenuURL       string   `json:"menu_url"`
	QrBannerText  string   `json:"qr_banner_text"`
	QrFooterItems []string `json:"qr_footer_items"`
}

type PlanResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TableResponse struct {
	ID           string `json:"id"`
	Number       int    `json:"number"`
	RestaurantID string `json:"restaurant_id"`
	QRCode       string `json:"qr_code"`
	IsActive     bool   `json:"is_active"`
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
	RequestID *string   `json:"request_id,omitempty"`
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

type UpdateAdminUserRequest struct {
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
}

type ResetPasswordResponse struct {
	NewPassword string `json:"new_password"`
}

type AdminUserResponse struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Role         string    `json:"role"`
	RestaurantID *string   `json:"restaurant_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// ─── Stats DTOs ───

type HourCountDTO struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

type HourScoreDTO struct {
	Hour          int     `json:"hour"`
	AvgScore      float64 `json:"avg_score"`
	FeedbackCount int     `json:"feedback_count"`
}

type StatsResponse struct {
	TotalTablesServed int            `json:"total_tables_served"`
	AvgServiceTimeSec float64        `json:"avg_service_time_seconds"`
	UsageByHour       []HourCountDTO `json:"usage_by_hour"`
	ScoreByHour       []HourScoreDTO `json:"score_by_hour"`
}

// ─── Mappers: Entity → Response ───

func ToRestaurantResponse(r *entity.Restaurant) RestaurantResponse {
	footer := []string(r.QrFooterItems)
	if footer == nil {
		footer = []string{}
	}
	planName := r.Plan.Name
	if planName == "" {
		planName = r.PlanID
	}
	return RestaurantResponse{
		ID:            r.ID,
		Name:          r.Name,
		PlanID:        r.PlanID,
		Plan:          planName,
		MenuURL:       r.MenuURL,
		QrBannerText:  r.QrBannerText,
		QrFooterItems: footer,
	}
}

func ToPlanResponse(p *entity.Plan) PlanResponse {
	return PlanResponse{ID: p.ID, Name: p.Name, Description: p.Description}
}

func ToPlanListResponse(plans []entity.Plan) []PlanResponse {
	out := make([]PlanResponse, len(plans))
	for i, p := range plans {
		out[i] = ToPlanResponse(&p)
	}
	return out
}

func ToTableResponse(t *entity.Table) TableResponse {
	return TableResponse{
		ID:           t.ID,
		Number:       t.Number,
		RestaurantID: t.RestaurantID,
		QRCode:       t.QRCode,
		IsActive:     t.IsActive,
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
		RequestID: f.RequestID,
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

func ToStatsResponse(s *usecase.RestaurantStats) StatsResponse {
	usage := make([]HourCountDTO, len(s.UsageByHour))
	for i, h := range s.UsageByHour {
		usage[i] = HourCountDTO{Hour: h.Hour, Count: h.Count}
	}
	scores := make([]HourScoreDTO, len(s.ScoreByHour))
	for i, h := range s.ScoreByHour {
		scores[i] = HourScoreDTO{Hour: h.Hour, AvgScore: h.AvgScore, FeedbackCount: h.Count}
	}
	return StatsResponse{
		TotalTablesServed: s.TotalTablesServed,
		AvgServiceTimeSec: s.AvgServiceTimeSec,
		UsageByHour:       usage,
		ScoreByHour:       scores,
	}
}

func ToAdminUserListResponse(users []entity.AdminUser) []AdminUserResponse {
	out := make([]AdminUserResponse, len(users))
	for i, u := range users {
		out[i] = ToAdminUserResponse(&u)
	}
	return out
}
