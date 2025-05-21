package service

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Viltsev/minishop/payment-service/internal/messaging"
	"github.com/Viltsev/minishop/payment-service/internal/model"
)

type PaymentService struct {
	store    model.PaymentStore
	rabbitMQ *messaging.RabbitMQ
}

func NewPaymentService(store model.PaymentStore, rabbitMQ *messaging.RabbitMQ) *PaymentService {
	return &PaymentService{
		store:    store,
		rabbitMQ: rabbitMQ,
	}
}

func (s *PaymentService) ProcessPayment(payment model.Payment, userService *UserServiceClient) (*model.Payment, error) {
	log.Printf("Пробуем снять средства")
	err := userService.Withdraw(payment.UserID, payment.Amount)
	if err != nil {
		payment.Status = "failed"
		s.store.CreatePayment(payment)

		event := map[string]interface{}{
			"type":    "PaymentFailed",
			"orderID": payment.OrderID,
			"userID":  payment.UserID,
			"email":   payment.Email,
			"amount":  payment.Amount,
			"error":   err.Error(),
		}
		body, _ := json.Marshal(event)
		s.rabbitMQ.Publish("payment.failed", body)
		log.Printf("Недостаточно средств")
		return nil, fmt.Errorf("failed to withdraw funds: %w", err)
	}

	// Деньги успешно списаны
	payment.Status = "completed"
	createdPayment, err := s.store.CreatePayment(payment)
	if err != nil {
		return nil, err
	}

	event := map[string]interface{}{
		"type":    "PaymentCompleted",
		"orderID": createdPayment.OrderID,
		"userID":  createdPayment.UserID,
		"email":   createdPayment.Email,
		"amount":  createdPayment.Amount,
	}
	body, _ := json.Marshal(event)
	s.rabbitMQ.Publish("payment.completed", body)

	log.Printf("Средства сняты")
	return createdPayment, nil
}

func (s *PaymentService) GetPaymentByID(id int) (*model.Payment, error) {
	return s.store.GetPaymentByID(id)
}

func (s *PaymentService) UpdateStatus(id int, status string) error {
	return s.store.UpdatePaymentStatus(id, status)
}

func (s *PaymentService) ListPaymentsByUser(userID int) ([]model.Payment, error) {
	return s.store.ListPaymentsByUser(userID)
}
