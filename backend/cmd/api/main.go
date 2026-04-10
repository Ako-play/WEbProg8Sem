package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"currencyparser/backend/internal/app"
	"currencyparser/backend/internal/config"
)

func main() {
	config.LoadDotenv()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("create app: %v", err)
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := application.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("starting %s on :%s", cfg.AppName, cfg.HTTPPort)
	if err := application.Start(); err != nil {
		log.Fatalf("start app: %v", err)
	}
}
