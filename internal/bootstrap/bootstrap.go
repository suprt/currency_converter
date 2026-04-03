package bootstrap

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/suprt/currency_converter/internal/client"
	"github.com/suprt/currency_converter/internal/config"
	"github.com/suprt/currency_converter/internal/handler"
	"github.com/suprt/currency_converter/internal/logger"
	"github.com/suprt/currency_converter/internal/middleware"
	"github.com/suprt/currency_converter/internal/repository"
	"github.com/suprt/currency_converter/internal/routes"
	"github.com/suprt/currency_converter/internal/service"
)

type App struct {
	Config      *config.Config
	Server      *http.Server
	Updater     *service.Updater
	RateLimiter routes.RateLimiter
	Repo        service.Repo
}

func NewApp() (*App, error) {
	cfg := config.Load()
	logger.Init(cfg.LogLevel)
	var repo service.Repo
	if cfg.RedisUse {
		redisRepo, err := repository.NewRedisStorage(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
		if err != nil {
			logger.Log.Error("failed to create redis storage", "error", err)
			return nil, err
		}
		repo = redisRepo
	} else {
		repo = repository.NewInMemoryRepository(cfg.InMemoryCleanupInterval)
	}

	apiClient := client.NewConverterClient(cfg.APIBaseURL, cfg.APIKey, cfg.ConverterTimeout)

	svc := service.NewConverterService(repo, apiClient)

	updater := service.NewUpdater(svc, cfg.UpdateInterval)
	updater.Start()

	rateLim := middleware.NewTokenBucket(middleware.LimiterConfig{
		RPS:   cfg.RPS,
		Burst: cfg.Burst,
	})
	rateLim.Start(cfg.RateLimiterCleanup)

	CacheHandler := handler.NewCacheHandler(svc)
	ConverterHandler := handler.NewConverterHandler(svc)
	CurrencyHandler := handler.NewCurrencyHandler(svc)
	HealthHandler := handler.NewHealthHandler(svc)

	router := routes.NewRouter(routes.RouterConfig{
		CacheOperator:       CacheHandler,
		CurrencyLister:      CurrencyHandler,
		HealthChecker:       HealthHandler,
		CurrencyConverter:   ConverterHandler,
		RateLimiter:         rateLim,
		APIKeyAuthenticator: middleware.APIKeyAuth(cfg.AdminAPIKey),
	})

	server := &http.Server{
		Addr:         cfg.ServerAddr(),
		Handler:      router,
		ReadTimeout:  cfg.ServerTimeout,
		WriteTimeout: cfg.ServerTimeout,
		IdleTimeout:  cfg.ServerIdleTimeout,
	}
	logger.Log.Info("server initialized", "addr", cfg.ServerAddr())
	return &App{
		Config:      cfg,
		Server:      server,
		Updater:     updater,
		RateLimiter: rateLim,
		Repo:        repo,
	}, nil
}

func (app *App) Stop() {

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err := app.Server.Shutdown(ctx)
	if err != nil {
		logger.Log.Error("server shutdown error", "error", err)
	}
	app.Updater.Stop()
	app.RateLimiter.Stop()
	if closer, ok := app.Repo.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			logger.Log.Error("repository close error", "error", err)
		}
	}
}
