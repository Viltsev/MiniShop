package app

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Viltsev/minishop/order-service/internal/handler"
	"github.com/Viltsev/minishop/order-service/internal/messaging"
	"github.com/Viltsev/minishop/order-service/internal/model"
	"github.com/Viltsev/minishop/order-service/internal/repository"
	"github.com/Viltsev/minishop/order-service/internal/service"
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

	orderStore := repository.NewStore(s.db)
	orderService := service.NewOrderService(orderStore, s.rabbitMQ)
	orderHandler := handler.NewOrderHandler(orderStore, *orderService)
	orderHandler.RegisterRoutes(subrouter)

	go func() {
		if err := s.startPaymentEventListener(orderStore); err != nil {
			log.Fatalf("failed to start order.created listener: %v", err)
		}
	}()

	log.Println("Server starts on http://localhost:8081")
	return http.ListenAndServe(s.addr, router)
}

func (s *APIServer) startPaymentEventListener(orderStore model.OrderStore) error {
	log.Println("[Listener] Initializing payment.* consumer...")

	return s.rabbitMQ.Consume("payment.*", func(body []byte) {
		log.Println("[Listener] Received payment event:", string(body))

		var event map[string]interface{}
		if err := json.Unmarshal(body, &event); err != nil {
			log.Printf("[Listener] Failed to unmarshal payment event: %v", err)
			return
		}

		eventType, ok := event["type"].(string)
		if !ok {
			log.Println("[Listener] Missing or invalid 'type' in event")
			return
		}

		orderIDFloat, ok := event["orderID"].(float64)
		if !ok {
			log.Println("[Listener] Missing or invalid 'orderID' in event")
			return
		}
		orderID := int(orderIDFloat)

		var newStatus string
		switch eventType {
		case "PaymentCompleted":
			newStatus = "completed"
		case "PaymentFailed":
			newStatus = "failed"
		default:
			log.Printf("[Listener] Unknown event type: %s", eventType)
			return
		}

		err := orderStore.UpdateStatus(orderID, newStatus)
		if err != nil {
			log.Printf("[Listener] Failed to update order status: %v", err)
		} else {
			log.Printf("[Listener] Order %d status updated to '%s'", orderID, newStatus)
		}
	})
}
