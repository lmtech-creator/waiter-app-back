package mocks

import (
	"fmt"

	"github.com/waiter/back/domain/entity"
)

// --- RestaurantRepository Mock ---

type RestaurantRepo struct {
	Restaurants map[string]*entity.Restaurant
	CreateFn    func(r *entity.Restaurant) error
	FindByIDFn  func(id string) (*entity.Restaurant, error)
	FindAllFn   func() ([]entity.Restaurant, error)
}

func NewRestaurantRepo() *RestaurantRepo {
	return &RestaurantRepo{Restaurants: make(map[string]*entity.Restaurant)}
}

func (m *RestaurantRepo) Create(r *entity.Restaurant) error {
	if m.CreateFn != nil {
		return m.CreateFn(r)
	}
	m.Restaurants[r.ID] = r
	return nil
}

func (m *RestaurantRepo) FindByID(id string) (*entity.Restaurant, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(id)
	}
	r, ok := m.Restaurants[id]
	if !ok {
		return nil, fmt.Errorf("restaurant not found")
	}
	return r, nil
}

func (m *RestaurantRepo) FindAll() ([]entity.Restaurant, error) {
	if m.FindAllFn != nil {
		return m.FindAllFn()
	}
	var list []entity.Restaurant
	for _, r := range m.Restaurants {
		list = append(list, *r)
	}
	return list, nil
}

// --- TableRepository Mock ---

type TableRepo struct {
	Tables                      map[string]*entity.Table
	CreateFn                    func(t *entity.Table) error
	FindByIDFn                  func(id string) (*entity.Table, error)
	FindByRestaurantFn          func(restaurantID string) ([]entity.Table, error)
	FindByNumberAndRestaurantFn func(number int, restaurantID string) (*entity.Table, error)
	FindByQRCodeFn              func(qrCode string) (*entity.Table, error)
	UpdateQRCodeFn              func(id, qrCode string) error
	InactiveTableFn             func(id string) error
	ReactiVateTableFn           func(id string) error
}

func NewTableRepo() *TableRepo {
	return &TableRepo{Tables: make(map[string]*entity.Table)}
}

func (m *TableRepo) Create(t *entity.Table) error {
	if m.CreateFn != nil {
		return m.CreateFn(t)
	}
	m.Tables[t.ID] = t
	return nil
}

func (m *TableRepo) FindByID(id string) (*entity.Table, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(id)
	}
	t, ok := m.Tables[id]
	if !ok {
		return nil, fmt.Errorf("table not found")
	}
	return t, nil
}

func (m *TableRepo) FindByRestaurantID(restaurantID string) ([]entity.Table, error) {
	if m.FindByRestaurantFn != nil {
		return m.FindByRestaurantFn(restaurantID)
	}
	var list []entity.Table
	for _, t := range m.Tables {
		if t.RestaurantID == restaurantID {
			list = append(list, *t)
		}
	}
	return list, nil
}

func (m *TableRepo) FindByNumberAndRestaurantID(number int, restaurantID string) (*entity.Table, error) {
	if m.FindByNumberAndRestaurantFn != nil {
		return m.FindByNumberAndRestaurantFn(number, restaurantID)
	}
	for _, t := range m.Tables {
		if t.Number == number && t.RestaurantID == restaurantID {
			return t, nil
		}
	}
	return nil, fmt.Errorf("table not found")
}

func (m *TableRepo) FindByQRCode(qrCode string) (*entity.Table, error) {
	if m.FindByQRCodeFn != nil {
		return m.FindByQRCodeFn(qrCode)
	}
	for _, t := range m.Tables {
		if t.QRCode == qrCode {
			return t, nil
		}
	}
	return nil, fmt.Errorf("table not found")
}

func (m *TableRepo) UpdateQRCode(id, qrCode string) error {
	if m.UpdateQRCodeFn != nil {
		return m.UpdateQRCodeFn(id, qrCode)
	}
	t, ok := m.Tables[id]
	if !ok {
		return fmt.Errorf("table not found")
	}
	t.QRCode = qrCode
	return nil
}

func (m *TableRepo) InactiveTable(id string) error {
	if m.InactiveTableFn != nil {
		return m.InactiveTableFn(id)
	}

	t, ok := m.Tables[id]
	if !ok {
		return fmt.Errorf("table not found")
	}

	t.IsActive = false
	return nil
}

func (m *TableRepo) ReactiVateTable(id string) error {
	if m.ReactiVateTableFn != nil {
		return m.ReactiVateTableFn(id)
	}

	t, ok := m.Tables[id]
	if !ok {
		return fmt.Errorf("table not found")
	}

	t.IsActive = true
	return nil
}

// --- RequestRepository Mock ---

type RequestRepo struct {
	Requests                 map[string]*entity.Request
	CreateFn                 func(r *entity.Request) error
	FindByIDFn               func(id string) (*entity.Request, error)
	FindActiveFn             func(restaurantID string) ([]entity.Request, error)
	FindByTableFn            func(tableID string) ([]entity.Request, error)
	FindLastCreatedByTableFn func(tableID string) (*entity.Request, error)
	UpdateStatusFn           func(id string, status entity.RequestStatus) error
}

func NewRequestRepo() *RequestRepo {
	return &RequestRepo{Requests: make(map[string]*entity.Request)}
}

func (m *RequestRepo) Create(r *entity.Request) error {
	if m.CreateFn != nil {
		return m.CreateFn(r)
	}
	m.Requests[r.ID] = r
	return nil
}

func (m *RequestRepo) FindByID(id string) (*entity.Request, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(id)
	}
	r, ok := m.Requests[id]
	if !ok {
		return nil, fmt.Errorf("request not found")
	}
	return r, nil
}

func (m *RequestRepo) FindActiveByRestaurantID(restaurantID string) ([]entity.Request, error) {
	if m.FindActiveFn != nil {
		return m.FindActiveFn(restaurantID)
	}
	var list []entity.Request
	for _, r := range m.Requests {
		if r.Status != entity.Done {
			list = append(list, *r)
		}
	}
	return list, nil
}

func (m *RequestRepo) FindByTableID(tableID string) ([]entity.Request, error) {
	if m.FindByTableFn != nil {
		return m.FindByTableFn(tableID)
	}
	var list []entity.Request
	for _, r := range m.Requests {
		if r.TableID == tableID {
			list = append(list, *r)
		}
	}
	return list, nil
}

func (m *RequestRepo) UpdateStatus(id string, status entity.RequestStatus) error {
	if m.UpdateStatusFn != nil {
		return m.UpdateStatusFn(id, status)
	}
	r, ok := m.Requests[id]
	if !ok {
		return fmt.Errorf("request not found")
	}
	r.Status = status
	return nil
}

func (m *RequestRepo) FindLastCreatedByTableID(tableID string) (*entity.Request, error) {
	if m.FindLastCreatedByTableFn != nil {
		return m.FindLastCreatedByTableFn(tableID)
	}
	var latest *entity.Request
	for _, r := range m.Requests {
		if r.TableID == tableID {
			if latest == nil || r.CreatedAt.After(latest.CreatedAt) {
				latest = r
			}
		}
	}
	if latest == nil {
		return nil, fmt.Errorf("request not found")
	}
	return latest, nil
}

// --- FeedbackRepository Mock ---

type FeedbackRepo struct {
	Feedbacks     map[string]*entity.Feedback
	CreateFn      func(f *entity.Feedback) error
	FindByTableFn func(tableID string) ([]entity.Feedback, error)
}

func NewFeedbackRepo() *FeedbackRepo {
	return &FeedbackRepo{Feedbacks: make(map[string]*entity.Feedback)}
}

func (m *FeedbackRepo) Create(f *entity.Feedback) error {
	if m.CreateFn != nil {
		return m.CreateFn(f)
	}
	m.Feedbacks[f.ID] = f
	return nil
}

func (m *FeedbackRepo) FindByTableID(tableID string) ([]entity.Feedback, error) {
	if m.FindByTableFn != nil {
		return m.FindByTableFn(tableID)
	}
	var list []entity.Feedback
	for _, f := range m.Feedbacks {
		if f.TableID == tableID {
			list = append(list, *f)
		}
	}
	return list, nil
}

// --- EventNotifier Mock ---

type Notifier struct {
	Events []NotifiedEvent
}

type NotifiedEvent struct {
	RestaurantID string
	Event        any
}

func NewNotifier() *Notifier {
	return &Notifier{}
}

func (m *Notifier) Notify(restaurantID string, event any) {
	m.Events = append(m.Events, NotifiedEvent{RestaurantID: restaurantID, Event: event})
}

// --- AdminRepository Mock ---

type AdminRepo struct {
	Admins             map[string]*entity.AdminUser
	FindByUsernameFn   func(username string) (*entity.AdminUser, error)
	FindByIDFn         func(id string) (*entity.AdminUser, error)
	FindByRestaurantFn func(restaurantID string) ([]entity.AdminUser, error)
	FindAllFn          func() ([]entity.AdminUser, error)
	CreateFn           func(admin *entity.AdminUser) error
	ExistsAnyFn        func() (bool, error)
	DeleteByIDFn       func(id string) error
}

func NewAdminRepo() *AdminRepo {
	return &AdminRepo{Admins: make(map[string]*entity.AdminUser)}
}

func (m *AdminRepo) FindByUsername(username string) (*entity.AdminUser, error) {
	if m.FindByUsernameFn != nil {
		return m.FindByUsernameFn(username)
	}
	for _, a := range m.Admins {
		if a.Username == username {
			return a, nil
		}
	}
	return nil, fmt.Errorf("admin not found")
}

func (m *AdminRepo) FindByID(id string) (*entity.AdminUser, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(id)
	}
	a, ok := m.Admins[id]
	if !ok {
		return nil, fmt.Errorf("admin not found")
	}
	return a, nil
}

func (m *AdminRepo) FindByRestaurantID(restaurantID string) ([]entity.AdminUser, error) {
	if m.FindByRestaurantFn != nil {
		return m.FindByRestaurantFn(restaurantID)
	}
	var list []entity.AdminUser
	for _, a := range m.Admins {
		if a.RestaurantID != nil && *a.RestaurantID == restaurantID {
			list = append(list, *a)
		}
	}
	return list, nil
}

func (m *AdminRepo) FindAll() ([]entity.AdminUser, error) {
	if m.FindAllFn != nil {
		return m.FindAllFn()
	}
	var list []entity.AdminUser
	for _, a := range m.Admins {
		list = append(list, *a)
	}
	return list, nil
}

func (m *AdminRepo) Create(admin *entity.AdminUser) error {
	if m.CreateFn != nil {
		return m.CreateFn(admin)
	}
	m.Admins[admin.ID] = admin
	return nil
}

func (m *AdminRepo) ExistsAny() (bool, error) {
	if m.ExistsAnyFn != nil {
		return m.ExistsAnyFn()
	}
	return len(m.Admins) > 0, nil
}

func (m *AdminRepo) DeleteByID(id string) error {
	if m.DeleteByIDFn != nil {
		return m.DeleteByIDFn(id)
	}
	delete(m.Admins, id)
	return nil
}
