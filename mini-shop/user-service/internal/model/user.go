package model

import (
	"time"
)

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
	GetUsers() ([]User, error)
	CreateUser(User) error
	DeleteUser(id int) error
	DeleteAllUsers() error
	UpdateUser(user *User) error
}

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
}
