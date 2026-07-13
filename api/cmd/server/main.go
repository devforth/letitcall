package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/letitcall/letitcall/api/internal/bootstrap"
	"github.com/letitcall/letitcall/api/internal/config"
	"github.com/letitcall/letitcall/api/internal/httpapi"
	"github.com/letitcall/letitcall/api/internal/store"
)

const (
	readTimeout     = 10 * time.Second
	writeTimeout    = 30 * time.Second
	idleTimeout     = 60 * time.Second
	shutdownTimeout = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	if err := config.LoadDotEnv(); err != nil {
		return fmt.Errorf("load environment files: %w", err)
	}
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	database, err := store.Open(cfg.Storage.LevelDBPath)
	if err != nil {
		return err
	}
	defer database.Close()

	if err := bootstrap.EnsureFirstUser(database, cfg.FirstUser, time.Now()); err != nil {
		return err
	}
	api, err := httpapi.New(cfg, database)
	if err != nil {
		return fmt.Errorf("create API server: %w", err)
	}
	signals, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		api.RunCalendarSync(signals)
	}()
	go func() {
		defer workers.Done()
		api.RunWebhookDelivery(signals)
	}()

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:           api.Handler(),
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    1 << 20,
	}

	serverError := make(chan error, 1)
	go func() {
		slog.Info("HTTP server listening", "port", cfg.HTTP.Port)
		serverError <- server.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		stop()
		workers.Wait()
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-signals.Done():
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	shutdownErr := server.Shutdown(shutdownContext)
	workers.Wait()
	if shutdownErr != nil {
		return fmt.Errorf("graceful shutdown: %w", shutdownErr)
	}
	return nil
}
