package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Viltsev/minishop/order-service/internal/model"
)

type OutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

// CreateOutboxEvent создает новую запись в таблице outbox
func (r *OutboxRepository) CreateOutboxEvent(ctx context.Context, event *model.OutboxEvent) error {
	query := `
        INSERT INTO outbox (event_type, payload, status, created_at, retry_count)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id`

	return r.db.QueryRowContext(
		ctx,
		query,
		event.EventType,
		event.Payload,
		event.Status,
		event.CreatedAt,
		event.RetryCount,
	).Scan(&event.ID)
}

// GetPendingEvents возвращает все pending события
func (r *OutboxRepository) GetPendingEvents(ctx context.Context) ([]*model.OutboxEvent, error) {
	query := `
        SELECT id, event_type, payload, status, created_at, processed_at, retry_count, error_message
        FROM outbox
        WHERE status = $1
        ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, model.OutboxStatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*model.OutboxEvent
	for rows.Next() {
		event := &model.OutboxEvent{}
		var processedAt sql.NullTime
		var errorMessage sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.Payload,
			&event.Status,
			&event.CreatedAt,
			&processedAt,
			&event.RetryCount,
			&errorMessage,
		)
		if err != nil {
			return nil, err
		}

		if processedAt.Valid {
			event.ProcessedAt = &processedAt.Time
		}
		if errorMessage.Valid {
			event.ErrorMessage = &errorMessage.String
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

// UpdateEventStatus обновляет статус события
func (r *OutboxRepository) UpdateEventStatus(ctx context.Context, eventID int, status string, errorMessage *string) error {
	query := `
        UPDATE outbox
        SET status = $1,
            processed_at = $2,
            error_message = $3,
            retry_count = retry_count + 1
        WHERE id = $4`

	processedAt := time.Now()
	_, err := r.db.ExecContext(ctx, query, status, processedAt, errorMessage, eventID)
	return err
}
