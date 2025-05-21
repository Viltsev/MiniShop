package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var Envs = LoadConfig()

type Config struct {
	Port                   string
	DBUser                 string
	DBPassword             string
	DBAddress              string
	DBName                 string
	JWTExpirationInSeconds int64
	JWTSecret              string
	SSLMode                string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Ошибка при загрузке .env файла")
	}

	cfg := &Config{
		JWTExpirationInSeconds: getEnvAsInt("JWT_EXP", 3600*24*7),
		JWTSecret:              getEnv("JWT_SECRET", "non-secret-anymore?"),
		Port:                   getEnv("DB_PORT", "5434"),
		DBUser:                 getEnv("DB_USER", "root"),
		DBPassword:             getEnv("DB_PASSWORD", ""),
		DBAddress:              getEnv("DB_HOST", "db"),
		DBName:                 getEnv("DB_NAME", "payment-service-db"),
		SSLMode:                getEnv("DB_SSL", "disable"),
	}

	return cfg
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	if defaultVal == "" {
		log.Fatalf("Ожидается переменная окружения: %s", key)
	}
	return defaultVal
}

func getEnvAsInt(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fallback
		}
		return i
	}
	return fallback
}
