package handler

import (
	"fmt"
	"mini-shop/user-service/internal/auth"
	"mini-shop/user-service/internal/config"
	"mini-shop/user-service/internal/model"
	"mini-shop/user-service/internal/service"
	"mini-shop/user-service/internal/utils"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	store          model.UserStore
	balanceService service.BalanceService
}

func NewUserHandler(store model.UserStore, balanceService service.BalanceService) *Handler {
	return &Handler{store: store, balanceService: balanceService}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/login", h.handleLogin).Methods("POST")
	router.HandleFunc("/register", h.handleRegister).Methods("POST")

	router.HandleFunc("/secret", auth.WithJWTAuth(h.secretMethod, h.store)).Methods("GET")

	router.HandleFunc("/users", auth.WithJWTAuth(h.getUsers, h.store)).Methods("GET")
	router.HandleFunc("/users/{id:[0-9]+}", auth.WithJWTAuth(h.deleteUser, h.store)).Methods("DELETE")
	router.HandleFunc("/users", auth.WithJWTAuth(h.deleteAllUsers, h.store)).Methods("DELETE")

	router.HandleFunc("/balance/{id:[0-9]+}", h.handleGetBalance).Methods("GET")
	router.HandleFunc("/balance/{id:[0-9]+}/add", h.handleAddBalance).Methods("POST")
	router.HandleFunc("/balance/{id:[0-9]+}/withdraw", h.handleWithdrawBalance).Methods("POST")
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var user model.LoginUserPayload
	if err := utils.ParseJSON(r, &user); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(user); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	u, err := h.store.GetUserByEmail(user.Email)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("not found, invalid email or password"))
		return
	}

	if !auth.ComparePasswords(u.Password, []byte(user.Password)) {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid email or password"))
		return
	}

	secret := []byte(config.Envs.JWTSecret)
	token, err := auth.CreateJWT(secret, u.ID, u.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var user model.RegisterUserPayload
	if err := utils.ParseJSON(r, &user); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(user); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	_, err := h.store.GetUserByEmail(user.Email)
	if err == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user with email %s already exists", user.Email))
		return
	}

	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.store.CreateUser(model.User{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Password:  hashedPassword,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	u, err := h.store.GetUserByEmail(user.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	secret := []byte(config.Envs.JWTSecret)
	accessToken, err := auth.CreateJWT(secret, u.ID, u.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	refreshToken, err := auth.CreateJWT(secret, u.ID, u.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]string{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func (h *Handler) secretMethod(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "секретный метод",
	})
}

func (h *Handler) getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.GetUsers()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to get users: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, users)
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	id, err := strconv.Atoi(userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user id"))
		return
	}

	err = h.store.DeleteUser(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to delete user: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
}

func (h *Handler) deleteAllUsers(w http.ResponseWriter, r *http.Request) {
	err := h.store.DeleteAllUsers()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to delete all users: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "all users deleted"})
}

func (h *Handler) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	balance, err := h.balanceService.GetBalance(id)
	if err != nil {
		utils.WriteError(w, http.StatusNotFound, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]float64{"balance": balance})
}

func (h *Handler) handleAddBalance(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	var payload struct {
		Amount float64 `json:"amount"`
	}

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if payload.Amount <= 0 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("amount must be positive"))
		return
	}

	balance, err := h.balanceService.AddBalance(id, payload.Amount)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]float64{"balance": balance})
}

func (h *Handler) handleWithdrawBalance(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid user ID"))
		return
	}

	var payload struct {
		Amount float64 `json:"amount"`
	}

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if payload.Amount <= 0 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("amount must be positive"))
		return
	}

	balance, err := h.balanceService.Withdraw(id, payload.Amount)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]float64{"balance": balance})
}
