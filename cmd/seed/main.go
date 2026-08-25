package main

import (
	"log"

	"github.com/dienulhaq/go-driving-course-management/config"
	"github.com/dienulhaq/go-driving-course-management/seeds"
)

func main() {
	cfg := config.Load()

	db, err := config.ConnectDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database initialization failed: %v", err)
	}

	seedConfig := seeds.Config{
		AdminName:     cfg.AdminName,
		AdminEmail:    cfg.AdminEmail,
		AdminPassword: cfg.AdminPassword,
	}
	if err := seeds.Run(db, seedConfig); err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	log.Println("seed completed")
}
