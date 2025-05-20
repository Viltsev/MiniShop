package model

import "time"

type OrderStore interface {
	CreateOrder(order Order) (*Order, error)
	GetOrderByID(id int) (*Order, error)
	UpdateStatus(id int, status string) error
	ListOrdersByUser(userID string) ([]Order, error)
	DeleteOrder(id int) error
}

type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userID"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type OrderRequest struct {
	Amount float64 `json:"amount" validate:"required"`
}
