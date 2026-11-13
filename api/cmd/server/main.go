package main

import (
	"context"
	"crypt-link/adapter/input/controller"
	"crypt-link/adapter/output/mongodb"
	"crypt-link/config"
	"crypt-link/core/service"
	"crypt-link/infra/database"
	"crypt-link/infra/router"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config.Load()
	slog.Info("configuration loaded")

	dbClient, err := database.Connect(ctx)
	if err != nil {
		return errors.Join(errors.New("failed to connect to database"), err)
	}
	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := dbClient.Disconnect(disconnectCtx); err != nil {
			slog.Error("failed to disconnect from database", "error", err)
		}
	}()
	slog.Info("database connected")

	// Dependency Injection (Hexagonal / Clean Architecture)
	repo := mongodb.NewMessageRepository(dbClient)
	svc := service.NewMessageService(repo)
	ctrl := controller.NewMessageController(svc)

	srv := &http.Server{
		Addr:              ":" + config.Port,
		Handler:           router.Setup(ctrl),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
