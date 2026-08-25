package main

import (
	"log"

	"github.com/dienulhaq/go-driving-course-management/config"
	_ "github.com/dienulhaq/go-driving-course-management/docs"
	"github.com/dienulhaq/go-driving-course-management/routes"
)

// @title Driving Course Management System API
// @version 1.0
// @description REST API Final Project Bootcamp Intensif Golang
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the Bearer prefix, for example: "Bearer eyJhbGci..."
// @securityDefinitions.basic BasicAuth
func main() {
	cfg := config.Load()

	db, err := config.ConnectDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database initialization failed: %v", err)
	}

	router, err := routes.New(db, cfg)
	if err != nil {
		log.Fatalf("router initialization failed: %v", err)
	}

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
