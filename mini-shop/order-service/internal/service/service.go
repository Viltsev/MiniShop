package service

import (
	"encoding/json"
	"fmt"

	"github.com/Viltsev/minishop/order-service/internal/messaging"
	"github.com/Viltsev/minishop/order-service/internal/model"
)

type OrderService struct {
	store    model.OrderStore
	rabbitMQ *messaging.RabbitMQ
}

func NewOrderService(store model.OrderStore, rabbitMQ *messaging.RabbitMQ) *OrderService {
	return &OrderService{
		store:    store,
		rabbitMQ: rabbitMQ,
	}
}

func (s *OrderService) CreateOrder(order model.Order) (*model.Order, error) {
	order.Status = "created"
	createdOrder, err := s.store.CreateOrder(order)
	if err != nil {
		return nil, err
	}

	// Публикуем событие OrderCreated
	event := map[string]interface{}{
		"type":    "OrderCreated",
		"orderID": createdOrder.ID,
		"userID":  createdOrder.UserID,
		"amount":  createdOrder.Amount,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	err = s.rabbitMQ.Publish("order.created", body)
	if err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return createdOrder, nil
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
