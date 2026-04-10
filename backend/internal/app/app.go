package app

import (
	"context"
	"fmt"
	"net/http"

	"currencyparser/backend/internal/config"
	"currencyparser/backend/internal/db"
	"currencyparser/backend/internal/repository"
	"currencyparser/backend/internal/service"
	httptransport "currencyparser/backend/internal/transport/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	cfg        config.Config
	pool       *pgxpool.Pool
	httpServer *http.Server
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.RunMigrations(ctx, pool, cfg.MigrationsDir); err != nil {
		pool.Close()
		return nil, err
	}

	authRepo := repository.NewAuthRepository(pool)
	ratesRepo := repository.NewRatesRepository(pool)
	authService := service.NewAuthService(authRepo, cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.AccessTTL, cfg.RefreshTTL)
	ratesService := service.NewRatesService(ratesRepo, cfg.ProviderBaseURL)
	handler := httptransport.NewHandler(authService, ratesService, cfg.FrontendOrigin)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	return &App{
		cfg:  cfg,
		pool: pool,
		httpServer: &http.Server{
			Addr:    ":" + cfg.HTTPPort,
			Handler: handler.Middleware(mux),
		},
	}, nil
}

func (a *App) Start() error {
	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	defer a.pool.Close()
	return a.httpServer.Shutdown(ctx)
}
