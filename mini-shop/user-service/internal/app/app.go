package app

import (
	"database/sql"
	"log"
	"net/http"

	"mini-shop/user-service/internal/handler"
	"mini-shop/user-service/internal/repository"
	"mini-shop/user-service/internal/service"

	"github.com/gorilla/mux"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {

	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	userStore := repository.NewStore(s.db)
	balanceService := service.NewBalanceService(userStore)
	userHandler := handler.NewUserHandler(userStore, *balanceService)
	userHandler.RegisterRoutes(subrouter)

	log.Println("Server starts on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
	return http.ListenAndServe(s.addr, router)
}
