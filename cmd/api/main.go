package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/suprt/currency_converter/internal/bootstrap"
)

func main() {
	app, err := bootstrap.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigint
		app.Stop()
		os.Exit(0)
	}()

	log.Printf("Server starting on %s", app.Config.ServerAddr())
	if err := app.Server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
