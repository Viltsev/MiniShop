package repository

import (
	"database/sql"
	"fmt"
	"mini-shop/user-service/internal/model"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func scanRowsIntoUser(rows *sql.Rows) (*model.User, error) {
	user := new(model.User)

	err := rows.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Store) GetUserByEmail(email string) (*model.User, error) {
	rows, err := s.db.Query("SELECT id, firstName, lastName, email, password, createdAt FROM users WHERE email = $1", email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	u := new(model.User)
	for rows.Next() {
		u, err = scanRowsIntoUser(rows)
		if err != nil {
			return nil, err
		}
	}

	if u.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return u, nil
}

func (s *Store) GetUserByID(id int) (*model.User, error) {
	rows, err := s.db.Query("SELECT id, firstName, lastName, email, password, createdAt FROM users WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	u := new(model.User)
	for rows.Next() {
		u, err = scanRowsIntoUser(rows)
		if err != nil {
			return nil, err
		}
	}

	if u.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return u, nil
}

func (s *Store) CreateUser(user model.User) error {
	_, err := s.db.Exec(
		"INSERT INTO users (firstName, lastName, email, password) VALUES ($1, $2, $3, $4)",
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetUsers() ([]model.User, error) {
	var users []model.User
	rows, err := s.db.Query("SELECT id, firstName, lastName, email, password, createdAt FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *Store) DeleteUser(id int) error {
	_, err := s.db.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

func (s *Store) DeleteAllUsers() error {
	_, err := s.db.Exec("DELETE FROM users")
	return err
}
