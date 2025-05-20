package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Viltsev/minishop/order-service/internal/auth"
	"github.com/Viltsev/minishop/order-service/internal/model"
	"github.com/Viltsev/minishop/order-service/internal/service"
	"github.com/Viltsev/minishop/order-service/internal/utils"
	"github.com/gorilla/mux"
)

type Handler struct {
	store   model.OrderStore
	service service.OrderService
}

func NewOrderHandler(store model.OrderStore, service service.OrderService) *Handler {
	return &Handler{store: store, service: service}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/orders", auth.WithJWTAuth(h.CreateOrder, h.store)).Methods("POST")
	router.HandleFunc("/orders/{id:[0-9]+}", h.GetOrder).Methods("GET")
	router.HandleFunc("/orders/{id:[0-9]+}/status", h.UpdateStatus).Methods("PUT")
	router.HandleFunc("/orders/user/{userID}", h.ListOrdersByUser).Methods("GET")
	router.HandleFunc("/orders/{id:[0-9]+}", h.DeleteOrder).Methods("DELETE")
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to create order")
	userID := r.Context().Value(auth.UserKey).(int)

	log.Println("user id ", userID)

	var orderRequest model.OrderRequest
	if err := utils.ParseJSON(r, &orderRequest); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	log.Println("Decoded order request:", orderRequest)

	order := model.Order{
		UserID: userID,
		Amount: orderRequest.Amount,
		Status: "created",
	}

	createdOrder, err := h.store.CreateOrder(order)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to create order: %v", err))
		return
	}

	log.Println("Order created successfully:", createdOrder)

	// Отправляем успешный ответ с созданным заказом
	utils.WriteJSON(w, http.StatusCreated, createdOrder)
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	order, err := h.service.GetOrderByID(id)
	if err != nil || order == nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(order)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	var payload struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateStatus(id, payload.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ListOrdersByUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]

	orders, err := h.service.ListOrdersByUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(orders)
}

func (h *Handler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	if err := h.service.DeleteOrder(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
