package app

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/Viltsev/minishop/order-service/internal/handler"
	"github.com/Viltsev/minishop/order-service/internal/messaging"
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

	log.Println("Server starts on http://localhost:8081")
	return http.ListenAndServe(s.addr, router)
}
