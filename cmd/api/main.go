package main

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/suprt/currency_converter/internal/bootstrap"
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
