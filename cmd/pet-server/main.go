package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zhanglei10281852-gif/pet-foster-go/internal/pet"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if len(os.Args) == 2 && os.Args[1] == "--healthcheck" {
		if err := healthcheck(); err != nil {
			os.Exit(1)
		}
		return
	}
	if err := run(logger); err != nil {
		logger.Error("pet server stopped", "error", err)
		os.Exit(1)
	}
}

func healthcheck() error {
	databasePath := os.Getenv("PET_DATABASE_PATH")
	if databasePath == "" {
		databasePath = "./data/pet-foster.db"
	}
	store, err := pet.Open(context.Background(), databasePath)
	if err != nil {
		return err
	}
	defer store.Close()
	return store.Ping(context.Background())
}

func run(logger *slog.Logger) error {
	ctx := context.Background()
	databasePath := os.Getenv("PET_DATABASE_PATH")
	if databasePath == "" {
		databasePath = "./data/pet-foster.db"
	}
	store, err := pet.Open(ctx, databasePath)
	if err != nil {
		return err
	}
	defer store.Close()
	service := pet.NewService(store)
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	server := &http.Server{Addr: addr, Handler: pet.NewHandler(service), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second}
	signalCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	errorsCh := make(chan error, 1)
	go func() { errorsCh <- server.ListenAndServe() }()
	select {
	case err := <-errorsCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-signalCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
	return nil
}
