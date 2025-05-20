package model

import "time"

// OrderCreatedEvent представляет событие создания заказа
type OrderCreatedEvent struct {
	OrderID   int       `json:"order_id"`
	UserID    int       `json:"user_id"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

// OutboxEvent представляет запись в таблице outbox
type OutboxEvent struct {
	ID           int        `json:"id"`
	EventType    string     `json:"event_type"`
	Payload      []byte     `json:"payload"` // JSON в виде []byte
	Status       string     `json:"status"`  // pending, processed, failed
	CreatedAt    time.Time  `json:"created_at"`
	ProcessedAt  *time.Time `json:"processed_at"` // может быть nil
	RetryCount   int        `json:"retry_count"`
	ErrorMessage *string    `json:"error_message"` // может быть nil
}

// Константы для статусов outbox
const (
	OutboxStatusPending   = "pending"
	OutboxStatusProcessed = "processed"
	OutboxStatusFailed    = "failed"
)

// Константы для типов событий
const (
	EventTypeOrderCreated = "OrderCreated"
)
