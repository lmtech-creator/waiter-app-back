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
	Tables             map[string]*entity.Table
	CreateFn           func(t *entity.Table) error
	FindByIDFn         func(id string) (*entity.Table, error)
	FindByRestaurantFn func(restaurantID string) ([]entity.Table, error)
	FindByQRCodeFn     func(qrCode string) (*entity.Table, error)
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

// --- RequestRepository Mock ---

type RequestRepo struct {
	Requests       map[string]*entity.Request
	CreateFn       func(r *entity.Request) error
	FindByIDFn     func(id string) (*entity.Request, error)
	FindActiveFn   func(restaurantID string) ([]entity.Request, error)
	FindByTableFn  func(tableID string) ([]entity.Request, error)
	UpdateStatusFn func(id string, status entity.RequestStatus) error
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
