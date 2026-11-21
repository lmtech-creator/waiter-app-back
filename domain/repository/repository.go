package repository

import (
	"time"

	"github.com/waiter/back/domain/entity"
)

type PlanRepository interface {
	Create(plan *entity.Plan) error
	FindByID(id string) (*entity.Plan, error)
	FindByName(name string) (*entity.Plan, error)
	FindAll() ([]entity.Plan, error)
	Update(plan *entity.Plan) error
	DeleteByID(id string) error
	CountRestaurants(planID string) (int64, error)
}

type RestaurantRepository interface {
	Create(restaurant *entity.Restaurant) error
	FindByID(id string) (*entity.Restaurant, error)
	FindAll() ([]entity.Restaurant, error)
	Update(restaurant *entity.Restaurant) error
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
	DeleteByID(id string) error
}

// HourCount is an aggregate bucket already expressed in the restaurant's
// reporting hour-of-day (UTC-3).
type HourCount struct {
	Hour  int
	Count int
}

// HourScore is an average score bucket already expressed in the restaurant's
// reporting hour-of-day (UTC-3). Count is the number of ratings behind the
// average, so callers can compute a weighted overall score.
type HourScore struct {
	Hour     int
	AvgScore float64
	Count    int
}

// HourBucket is a raw UTC hour bucket returned by the database. The use case
// converts it to the restaurant's local hour so the offset never depends on the
// Postgres session timezone.
type HourBucket struct {
	Start time.Time
	Count int
}

// ScoreBucket is a raw UTC hour bucket with the sum and count needed to compute
// a weighted average after merging buckets into local hours.
type ScoreBucket struct {
	Start    time.Time
	ScoreSum float64
	Count    int
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
	UsageByHour(restaurantID string, since, until time.Time) ([]HourBucket, error)
}

type FeedbackRepository interface {
	Create(feedback *entity.Feedback) error
	FindByTableID(tableID string) ([]entity.Feedback, error)
	FindByRestaurantID(restaurantID string) ([]entity.Feedback, error)
	AvgScoreByHour(restaurantID string, since, until time.Time) ([]ScoreBucket, error)
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
