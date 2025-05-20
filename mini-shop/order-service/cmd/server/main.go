package main

import (
	"database/sql"
	"log"

	"github.com/Viltsev/minishop/order-service/internal/app"
	"github.com/Viltsev/minishop/order-service/internal/config"
	"github.com/Viltsev/minishop/order-service/internal/database"
	"github.com/Viltsev/minishop/order-service/internal/messaging"
)

func main() {
	log.Print("START ORDER SERVICE")

	cfg := database.Config{
		Host:     config.Envs.DBAddress,
		Port:     config.Envs.Port,
		User:     config.Envs.DBUser,
		Password: config.Envs.DBPassword,
		DBName:   config.Envs.DBName,
		SSLMode:  config.Envs.SSLMode,
	}

	log.Println("Connecting to DB...")
	db, err := database.NewPostgresStorage(cfg)
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}
	log.Println("DB connection established")

	initStorage(db)

	log.Println("Running migrations...")
	if err := database.RunMigrations(cfg); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("Migrations applied successfully.")

	log.Println("Connecting to RabbitMQ...")
	rabbitURL := "amqp://guest:guest@rabbitmq:5672/"
	rabbitMQ, err := messaging.NewRabbitMQ(rabbitURL, "orders_exchange")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer rabbitMQ.Close()
	log.Println("Connected to RabbitMQ")

	log.Println("Starting API server...")
	server := app.NewAPIServer(":8081", db, rabbitMQ)
	if err := server.Run(); err != nil {
		log.Fatal("API server failed:", err)
	}

	log.Print("server has started")
}

func initStorage(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Println("Waiting for DB to be ready...")
	}
	log.Println("DB is ready!")
}
