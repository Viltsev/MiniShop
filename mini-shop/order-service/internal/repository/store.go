package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Viltsev/minishop/order-service/internal/model"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func scanRowsIntoOrder(rows *sql.Rows) (*model.Order, error) {
	user := new(model.Order)

	err := rows.Scan(
		&user.ID,
		&user.UserID,
		&user.Amount,
		&user.Status,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Store) CreateOrder(order model.Order) (*model.Order, error) {
	query := `INSERT INTO orders (userID, amount, status, createdAt) VALUES ($1, $2, $3, $4) RETURNING id`
	err := s.db.QueryRow(query, order.UserID, order.Amount, order.Status, time.Now()).Scan(&order.ID)
	if err != nil {
		return nil, err
	}
	order.CreatedAt = time.Now()
	return &order, nil
}

func (s *Store) GetOrderByID(id int) (*model.Order, error) {
	query := `SELECT id, userID, amount, status, createdAt FROM orders WHERE id = $1`

	row := s.db.QueryRow(query, id)

	order := &model.Order{}
	err := row.Scan(&order.ID, &order.UserID, &order.Amount, &order.Status, &order.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return order, nil
}

func (s *Store) UpdateStatus(id int, status string) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`

	result, err := s.db.Exec(query, status, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no order found with ID %d", id)
	}

	return nil
}

func (s *Store) ListOrdersByUser(userID string) ([]model.Order, error) {
	query := `SELECT id, userID, amount, status, createdAt FROM orders WHERE userID = $1`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		order, err := scanRowsIntoOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *order)
	}

	return orders, nil
}

func (s *Store) DeleteOrder(id int) error {
	query := `DELETE FROM orders WHERE id = $1`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no order found with ID %d", id)
	}

	return nil
}
