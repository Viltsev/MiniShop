package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Viltsev/minishop/payment-service/internal/handler"
	"github.com/Viltsev/minishop/payment-service/internal/messaging"
	"github.com/Viltsev/minishop/payment-service/internal/model"
	"github.com/Viltsev/minishop/payment-service/internal/repository"
	"github.com/Viltsev/minishop/payment-service/internal/service"
	"github.com/gorilla/mux"
)

type APIServer struct {
	addr     string
	db       *sql.DB
	rabbitMQ *messaging.RabbitMQ
}

func NewAPIServer(addr string, db *sql.DB, rabbitMQ *messaging.RabbitMQ) *APIServer {
	return &APIServer{
		addr:     addr,
		db:       db,
		rabbitMQ: rabbitMQ,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	paymentStore := repository.NewStore(s.db)

	paymentService := service.NewPaymentService(paymentStore, s.rabbitMQ)
	paymentHandler := handler.NewPaymentHandler(paymentStore, paymentService)
	paymentHandler.RegisterRoutes(subrouter)

	go func() {
		if err := s.startOrderCreatedListener(paymentService); err != nil {
			log.Fatalf("failed to start order.created listener: %v", err)
		}
	}()

	log.Println("Server starts on http://localhost:8082")
	return http.ListenAndServe(s.addr, router)
}

// startOrderCreatedListener подписывается на очередь "order.created" и обрабатывает события создания заказа
func (s *APIServer) startOrderCreatedListener(paymentService *service.PaymentService) error {
	log.Println("[Listener] Initializing order.created consumer...")
	return s.rabbitMQ.Consume("order.created", func(body []byte) {
		log.Println("[Listener] Received raw order.created event:", string(body))

		var orderEvent model.OrderCreatedEvent
		if err := json.Unmarshal(body, &orderEvent); err != nil {
			log.Printf("[Listener] Failed to unmarshal order event: %v", err)
			return
		}

		log.Printf("[Listener] Parsed order event: %+v", orderEvent)

		payment := model.Payment{
			OrderID: orderEvent.OrderID,
			UserID:  orderEvent.UserID,
			Email:   orderEvent.Email,
			Amount:  orderEvent.Amount,
			Status:  "pending",
		}

		url := fmt.Sprintf("http://user-service:8080/api/v1/balance/%d", orderEvent.UserID)
		userServiceClient := service.NewUserServiceClient(url)

		_, err := paymentService.ProcessPayment(payment, userServiceClient)
		if err != nil {
			log.Printf("[Listener] Failed to process payment: %v", err)
		} else {
			log.Println("[Listener] Payment processed successfully for order:", orderEvent.OrderID)
		}
	})
}
