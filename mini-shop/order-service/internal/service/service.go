package service

import "github.com/Viltsev/minishop/order-service/internal/model"

type OrderService struct {
	store model.OrderStore
}

func NewOrderService(store model.OrderStore) *OrderService {
	return &OrderService{store: store}
}

func (s *OrderService) CreateOrder(order model.Order) (*model.Order, error) {
	order.Status = "pending"
	return s.store.CreateOrder(order)
}

func (s *OrderService) GetOrderByID(id int) (*model.Order, error) {
	return s.store.GetOrderByID(id)
}

func (s *OrderService) UpdateStatus(id int, status string) error {
	return s.store.UpdateStatus(id, status)
}

func (s *OrderService) ListOrdersByUser(userID string) ([]model.Order, error) {
	return s.store.ListOrdersByUser(userID)
}

func (s *OrderService) DeleteOrder(id int) error {
	return s.store.DeleteOrder(id)
}
