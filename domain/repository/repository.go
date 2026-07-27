package repository

import (
	"time"

	"github.com/waiter/back/domain/entity"
)

type RestaurantRepository interface {
	Create(restaurant *entity.Restaurant) error
	FindByID(id string) (*entity.Restaurant, error)
	FindAll() ([]entity.Restaurant, error)
}

type TableRepository interface {
	Create(table *entity.Table) error
	FindByID(id string) (*entity.Table, error)
	FindByRestaurantID(restaurantID string) ([]entity.Table, error)
	FindByNumberAndRestaurantID(number int, restaurantID string) (*entity.Table, error)
	FindByQRCode(qrCode string) (*entity.Table, error)
	InactiveTable(id string) error
	ReactiVateTable(id string) error
	UpdateQRCode(id, qrCode string) error
}

type HourCount struct {
	Hour  int
	Count int
}

type HourScore struct {
	Hour     int
	AvgScore float64
}

type RequestRepository interface {
	Create(request *entity.Request) error
	FindByID(id string) (*entity.Request, error)
	FindActiveByRestaurantID(restaurantID string) ([]entity.Request, error)
	FindByTableID(tableID string) ([]entity.Request, error)
	FindLastCreatedByTableID(tableID string) (*entity.Request, error)
	UpdateStatus(id string, status entity.RequestStatus) error
	UpdateCompletedAt(id string, t time.Time) error
	CountServedTables(restaurantID string, since, until time.Time) (int, error)
	AvgServiceTime(restaurantID string, since, until time.Time) (float64, error)
	UsageByHour(restaurantID string, since, until time.Time) ([]HourCount, error)
}

type FeedbackRepository interface {
	Create(feedback *entity.Feedback) error
	FindByTableID(tableID string) ([]entity.Feedback, error)
	FindByRestaurantID(restaurantID string) ([]entity.Feedback, error)
	AvgScoreByHour(restaurantID string, since, until time.Time) ([]HourScore, error)
}

type AdminRepository interface {
	FindByUsername(username string) (*entity.AdminUser, error)
	FindByID(id string) (*entity.AdminUser, error)
	FindByRestaurantID(restaurantID string) ([]entity.AdminUser, error)
	FindAll() ([]entity.AdminUser, error)
	Create(admin *entity.AdminUser) error
	Update(admin *entity.AdminUser) error
	ExistsAny() (bool, error)
	DeleteByID(id string) error
}
