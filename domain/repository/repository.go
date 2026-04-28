package repository

import "github.com/waiter/back/domain/entity"

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
	UpdateQRCode(id, qrCode string) error
}

type RequestRepository interface {
	Create(request *entity.Request) error
	FindByID(id string) (*entity.Request, error)
	FindActiveByRestaurantID(restaurantID string) ([]entity.Request, error)
	FindByTableID(tableID string) ([]entity.Request, error)
	FindLastCreatedByTableID(tableID string) (*entity.Request, error)
	UpdateStatus(id string, status entity.RequestStatus) error
}

type FeedbackRepository interface {
	Create(feedback *entity.Feedback) error
	FindByTableID(tableID string) ([]entity.Feedback, error)
}

type AdminRepository interface {
	FindByUsername(username string) (*entity.AdminUser, error)
	Create(admin *entity.AdminUser) error
	ExistsAny() (bool, error)
}
