package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSONB is a slice of strings that marshals to/from a JSONB column.
type JSONB []string

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONB) Scan(src interface{}) error {
	if src == nil {
		*j = nil
		return nil
	}
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("JSONB.Scan: expected []byte, got %T", src)
	}
	return json.Unmarshal(bytes, j)
}

type Restaurant struct {
	ID            string   `json:"id" gorm:"primaryKey;type:uuid"`
	Name          string   `json:"name" gorm:"not null"`
	Plan          string   `json:"plan" gorm:"not null;default:free"`
	QrBannerText  string   `json:"qr_banner_text" gorm:"not null;default:''"`
	QrFooterItems JSONB    `json:"qr_footer_items" gorm:"type:jsonb;not null;default:'[]'"`
}

type Table struct {
	ID           string `json:"id" gorm:"primaryKey;type:uuid"`
	Number       int    `json:"number" gorm:"not null"`
	RestaurantID string `json:"restaurant_id" gorm:"type:uuid;not null;index"`
	QRCode       string `json:"qr_code" gorm:"not null;uniqueIndex"`
	IsActive     bool   `json:"is_active" gorm:"not null;default:true"`
}

type RequestType string

const (
	CallWaiter RequestType = "CALL_WAITER"
	AskBill    RequestType = "ASK_BILL"
	AskHelp    RequestType = "ASK_HELP"
)

type RequestStatus string

const (
	Pending   RequestStatus = "PENDING"
	InProcess RequestStatus = "IN_PROCESS"
	Done      RequestStatus = "DONE"
)

type Request struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid"`
	TableID     string         `json:"table_id" gorm:"type:uuid;not null;index"`
	Type        RequestType    `json:"type" gorm:"type:varchar(20);not null"`
	Status      RequestStatus  `json:"status" gorm:"type:varchar(20);not null;default:PENDING"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	CompletedAt *time.Time     `json:"completed_at,omitempty" gorm:"index"`
}

type Feedback struct {
	ID        string     `json:"id" gorm:"primaryKey;type:uuid"`
	TableID   string     `json:"table_id" gorm:"type:uuid;not null;index"`
	RequestID *string    `json:"request_id,omitempty" gorm:"type:uuid"`
	Score     int        `json:"score" gorm:"not null"`
	Comment   string     `json:"comment,omitempty"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

type AdminRole string

const (
	RoleSuperAdmin AdminRole = "superadmin"
	RoleOwner      AdminRole = "owner"
	RoleEmployee   AdminRole = "employee"
)

type AdminUser struct {
	ID           string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	RestaurantID *string   `json:"restaurant_id,omitempty" gorm:"type:uuid;index"`
	Username     string    `json:"username" gorm:"not null;unique"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Role         AdminRole `json:"role" gorm:"type:varchar(20);not null;default:employee"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}
