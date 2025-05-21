package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Viltsev/minishop/payment-service/internal/config"
	"github.com/Viltsev/minishop/payment-service/internal/model"
	"github.com/Viltsev/minishop/payment-service/internal/utils"
	"github.com/golang-jwt/jwt"
)

type contextKey string

const UserKey contextKey = "userID"

func WithJWTAuth(handlerFunc http.HandlerFunc, store model.PaymentStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := utils.GetTokenFromRequest(r)

		userID, err := GetUserIDFromToken(tokenString)
		if err != nil {
			log.Printf("failed to validate token: %v", err)
			permissionDenied(w)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserKey, userID)
		r = r.WithContext(ctx)

		handlerFunc(w, r)
	}
}

func GetUserIDFromToken(tokenString string) (int, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := validateJWT(tokenString)
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid token claims")
	}

	userIDStr, ok := claims["userID"].(string)
	if !ok {
		return 0, fmt.Errorf("userID not found in token claims")
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, fmt.Errorf("failed to convert userID to int: %w", err)
	}

	return userID, nil
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Envs.JWTSecret), nil
	})
}

func permissionDenied(w http.ResponseWriter) {
	utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied"))
}
