package handler

import (
	"net/http"
	"strconv"

	"github.com/Viltsev/minishop/payment-service/internal/auth"
	"github.com/Viltsev/minishop/payment-service/internal/model"
	"github.com/Viltsev/minishop/payment-service/internal/service"
	"github.com/Viltsev/minishop/payment-service/internal/utils"
	"github.com/gorilla/mux"
)

type Handler struct {
	store   model.PaymentStore
	service *service.PaymentService
}

func NewPaymentHandler(store model.PaymentStore, service *service.PaymentService) *Handler {
	return &Handler{store: store, service: service}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Защищенный маршрут, где user может получить историю платежей
	router.HandleFunc("/payments/user", auth.WithJWTAuth(h.ListPaymentsByUser, h.store)).Methods("GET")
	// Получение конкретного платежа (можно защитить или оставить открытым, по желанию)
	router.HandleFunc("/payments/{id:[0-9]+}", h.GetPaymentByID).Methods("GET")
}

func (h *Handler) ListPaymentsByUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserKey).(int)

	payments, err := h.service.ListPaymentsByUser(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, payments)
}

func (h *Handler) GetPaymentByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	payment, err := h.service.GetPaymentByID(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if payment == nil {
		http.NotFound(w, r)
		return
	}

	utils.WriteJSON(w, http.StatusOK, payment)
}
