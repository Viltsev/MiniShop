package service

import (
	"fmt"
	"log"
	"mini-shop/user-service/internal/model"
)

type BalanceService struct {
	store model.UserStore
}

func NewBalanceService(store model.UserStore) *BalanceService {
	return &BalanceService{
		store: store,
	}
}

func (s *BalanceService) GetBalance(userID int) (float64, error) {
	user, err := s.store.GetUserByID(userID)
	if err != nil {
		log.Printf("Error getting user by ID %d: %v", userID, err)
		return 0, fmt.Errorf("user not found")
	}
	return user.Balance, nil
}

func (s *BalanceService) AddBalance(userID int, amount float64) (float64, error) {
	user, err := s.store.GetUserByID(userID)
	if err != nil {
		log.Printf("Error getting user by ID %d: %v", userID, err)
		return 0, fmt.Errorf("user not found")
	}

	if amount < 0 {
		return 0, fmt.Errorf("cannot add a negative amount")
	}

	user.Balance += amount
	err = s.store.UpdateUser(user)
	if err != nil {
		log.Printf("Error updating user balance for user ID %d: %v", userID, err)
		return 0, fmt.Errorf("failed to update balance")
	}

	return user.Balance, nil
}

func (s *BalanceService) Withdraw(userID int, amount float64) (float64, error) {
	user, err := s.store.GetUserByID(userID)
	if err != nil {
		log.Printf("Error getting user by ID %d: %v", userID, err)
		return 0, fmt.Errorf("user not found")
	}

	if user.Balance < amount {
		return 0, fmt.Errorf("insufficient funds")
	}

	user.Balance -= amount
	err = s.store.UpdateUser(user)
	if err != nil {
		log.Printf("Error updating user balance for user ID %d: %v", userID, err)
		return 0, fmt.Errorf("failed to update balance")
	}

	return user.Balance, nil
}
