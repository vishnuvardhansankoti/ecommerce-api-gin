package main

import (
	"context"
	"log"

	_ "ecommerce-api-gin/docs"
	"ecommerce-api-gin/internal/config"
	"ecommerce-api-gin/internal/database"
	"ecommerce-api-gin/internal/router"
)

// @title E-commerce API
// @version 1.0
// @description Sample REST API for an e-commerce portal built with Gin and PostgreSQL.
// @BasePath /
// @schemes http https
func main() {
	cfg := config.Load()

	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(context.Background(), db); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	r := router.New(db)

	log.Printf("server listening on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
