// @title Currency Converter API
// @version 1.0
// @description A REST API for converting currencies with caching support
// @description
// @description ## Features
// @description - Real-time currency conversion
// @description - Caching with Redis or in-memory storage
// @description - Rate limiting
// @description - Admin endpoints for cache management
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/suprt/currency_converter

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY
// @description Admin API key for accessing cache management endpoints
package main

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/suprt/currency_converter/internal/bootstrap"
	_ "github.com/suprt/currency_converter/internal/handler/docs"
	"github.com/suprt/currency_converter/internal/logger"
)

func main() {
	app, err := bootstrap.NewApp()
	if err != nil {
		logger.Error("failed to initialize app", "error", err)
		os.Exit(1)
	}
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigint
		app.Stop()
		os.Exit(0)
	}()

	logger.Info("server starting", "addr", app.Config.ServerAddr())
	if err := app.Server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
