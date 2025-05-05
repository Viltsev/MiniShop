package main

import (
	"database/sql"
	"log"
	"mini-shop/user-service/internal/app"
	"mini-shop/user-service/internal/config"
	"mini-shop/user-service/internal/database"
)

func main() {
	log.Println("STARTING USER SERVICE")
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
		log.Fatal(err)
		log.Fatal("DB connection failed:", err)
	}
	log.Println("DB connection established")

	initStorage(db)

	log.Println("Running migrations...")
	if err := database.RunMigrations(cfg); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("Migrations applied successfully.")

	log.Println("Starting API server...")
	server := app.NewAPIServer(":8080", db)
	if err := server.Run(); err != nil {
		log.Fatal(err)
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
