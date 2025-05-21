package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type UserServiceClient struct {
	BaseURL string
}

func NewUserServiceClient(baseURL string) *UserServiceClient {
	return &UserServiceClient{BaseURL: baseURL}
}

func (u *UserServiceClient) Withdraw(userID int, amount float64) error {
	payload := map[string]interface{}{
		"amount": amount,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(fmt.Sprintf("%s/withdraw", u.BaseURL), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to contact user service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("withdrawal failed with status: %d", resp.StatusCode)
	}

	return nil
}
