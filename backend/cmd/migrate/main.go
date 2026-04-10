package main

import (
	"context"
	"log"

	"currencyparser/backend/internal/config"
	"currencyparser/backend/internal/db"
)

func main() {
	config.LoadDotenv()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()
	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool, cfg.MigrationsDir); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	log.Println("migrations completed")
}
