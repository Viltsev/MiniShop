package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Viltsev/minishop/payment-service/internal/model"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func scanRowsIntoPayment(rows *sql.Rows) (*model.Payment, error) {
	payment := new(model.Payment)

	err := rows.Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.UserID,
		&payment.Amount,
		&payment.Status,
		&payment.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *Store) CreatePayment(payment model.Payment) (*model.Payment, error) {
	query := `INSERT INTO payments (orderID, userID, amount, status, createdAt) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	now := time.Now()
	err := s.db.QueryRow(query, payment.OrderID, payment.UserID, payment.Amount, payment.Status, now).Scan(&payment.ID)
	if err != nil {
		return nil, err
	}
	payment.CreatedAt = now
	return &payment, nil
}

func (s *Store) GetPaymentByID(id int) (*model.Payment, error) {
	query := `SELECT id, orderID, userID, amount, status, createdAt FROM payments WHERE id = $1`

	row := s.db.QueryRow(query, id)

	payment := &model.Payment{}
	err := row.Scan(&payment.ID, &payment.OrderID, &payment.UserID, &payment.Amount, &payment.Status, &payment.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return payment, nil
}

func (s *Store) UpdatePaymentStatus(id int, status string) error {
	query := `UPDATE payments SET status = $1 WHERE id = $2`

	result, err := s.db.Exec(query, status, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no payment found with ID %d", id)
	}

	return nil
}

func (s *Store) ListPaymentsByUser(userID int) ([]model.Payment, error) {
	query := `SELECT id, orderID, userID, amount, status, createdAt FROM payments WHERE userID = $1`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []model.Payment
	for rows.Next() {
		payment, err := scanRowsIntoPayment(rows)
		if err != nil {
			return nil, err
		}
		payments = append(payments, *payment)
	}

	return payments, nil
}
