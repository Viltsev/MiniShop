package app

import (
	"encoding/json"
	"log"

	"github.com/Viltsev/notification-service/internal/messaging"
	"github.com/Viltsev/notification-service/internal/service"
)

type APIServer struct {
	addr     string
	rabbitMQ *messaging.RabbitMQ
}

func NewAPIServer(addr string, rabbitMQ *messaging.RabbitMQ) *APIServer {
	return &APIServer{
		addr:     addr,
		rabbitMQ: rabbitMQ,
	}
}

func (s *APIServer) Run() error {
	notificationService := service.NewNotificationService(s.rabbitMQ)

	go func() {
		if err := s.startPaymentEventListener(notificationService); err != nil {
			log.Fatalf("failed to start notification service listener: %v", err)
		}
	}()

	log.Println("Notification server starts http://localhost:8083")
	return nil
}

func (s *APIServer) startPaymentEventListener(notificationService *service.NotificationService) error {
	log.Println("[Listener-NS] Initializing payment.* consumer...")

	return s.rabbitMQ.Consume("payment.*", func(body []byte) {
		log.Println("[Listener-NS] Received payment event:", string(body))

		var event map[string]interface{}
		if err := json.Unmarshal(body, &event); err != nil {
			log.Printf("[Listener-NS] Failed to unmarshal payment event: %v", err)
			return
		}

		notificationService.HandlePaymentEvent(body)
	})
}
