package model

import "time"

type PaymentStore interface {
	CreatePayment(payment Payment) (*Payment, error)
	GetPaymentByID(id int) (*Payment, error)
	UpdatePaymentStatus(id int, status string) error
	ListPaymentsByUser(userID int) ([]Payment, error)
}

type Payment struct {
	ID        int       `db:"id"`
	OrderID   int       `db:"order_id"`
	UserID    int       `db:"user_id"`
	Amount    float64   `db:"amount"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
}

type OutboxEvent struct {
	ID        int        `db:"id"`
	EventType string     `db:"event_type"`
	Payload   []byte     `db:"payload"`
	Status    string     `db:"status"`
	CreatedAt time.Time  `db:"created_at"`
	SentAt    *time.Time `db:"sent_at"`
}

type OrderCreatedEvent struct {
	OrderID int     `json:"orderID"`
	UserID  int     `json:"userID"`
	Amount  float64 `json:"amount"`
}
