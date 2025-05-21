package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Viltsev/notification-service/internal/app"
	"github.com/Viltsev/notification-service/internal/messaging"
)

func main() {
	log.Println("START NOTIFICATION SERVICE")

	rabbitURL := "amqp://guest:guest@rabbitmq:5672/"
	rabbitMQ, err := messaging.NewRabbitMQ(rabbitURL, "minishop")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer rabbitMQ.Close()
	log.Println("Connected to RabbitMQ")

	server := app.NewAPIServer(":8083", rabbitMQ)

	if err := server.Run(); err != nil {
		log.Fatalf("API server failed to start: %v", err)
	}

	log.Println("Notification server has started")

	// Отлавливаем сигналы ОС для корректного завершения
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Ждём сигнала завершения
	sig := <-sigs
	log.Printf("Received signal %s, shutting down...", sig)
	log.Println("Notification service stopped")
}
