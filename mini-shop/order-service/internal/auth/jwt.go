package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Viltsev/minishop/order-service/internal/config"
	"github.com/Viltsev/minishop/order-service/internal/model"
	"github.com/Viltsev/minishop/order-service/internal/utils"
	"github.com/golang-jwt/jwt"
)

type contextKey string

const (
	UserKey  contextKey = "userID"
	EmailKey contextKey = "email"
)

func WithJWTAuth(handlerFunc http.HandlerFunc, store model.OrderStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем токен из заголовка Authorization
		tokenString := utils.GetTokenFromRequest(r)

		// Валидируем токен и извлекаем userID
		userID, email, err := GetUserIDFromToken(tokenString)
		if err != nil {
			log.Printf("failed to validate token: %v", err)
			permissionDenied(w)
			return
		}

		// Сохраняем userID в контексте запроса
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserKey, userID)
		ctx = context.WithValue(ctx, EmailKey, email)
		r = r.WithContext(ctx)

		// Вызываем основной обработчик
		handlerFunc(w, r)
	}
}

// GetUserIDFromToken извлекает userID из JWT токена
func GetUserIDFromToken(tokenString string) (int, string, error) {
	// Проверяем, начинается ли токен с "Bearer " и удаляем это слово, если оно есть
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Валидируем и парсим токен
	token, err := validateJWT(tokenString)
	if err != nil {
		log.Printf("failed to validate token: %v", err)
		return 0, "", fmt.Errorf("invalid token: %w", err)
	}

	// Проверяем валидность токена
	if !token.Valid {
		log.Println("invalid token")
		return 0, "", fmt.Errorf("invalid token")
	}

	// Извлекаем данные из токена (claims)
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", fmt.Errorf("invalid token claims")
	}

	// Получаем userID из claims
	userIDStr, ok := claims["userID"].(string)
	if !ok {
		return 0, "", fmt.Errorf("userID not found in token claims")
	}

	// Получение email из claims
	email, ok := claims["email"].(string)
	if !ok {
		return 0, "", fmt.Errorf("email not found in token claims")
	}

	// Преобразуем userID в целое число
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, "", fmt.Errorf("failed to convert userID to int: %w", err)
	}

	return userID, email, nil
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	// Функция для проверки токена
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что метод подписи правильный
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Возвращаем секрет для верификации
		return []byte(config.Envs.JWTSecret), nil
	})
}

func permissionDenied(w http.ResponseWriter) {
	// Возвращаем ошибку доступа
	utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied"))
}
