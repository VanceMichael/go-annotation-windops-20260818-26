package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	"windops/internal/application"
	"windops/internal/config"
	"windops/internal/httpapi"
	"windops/internal/platform/clock"
	"windops/internal/platform/identity"
	"windops/internal/store"
	"windops/internal/worker"
)

func main() {
	if err := run(); err != nil {
		slog.Error("windops stopped", "error", err)
		os.Exit(1)
	}
}
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if dir := filepath.Dir(cfg.DatabaseDSN); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	db, err := store.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	defer db.Close()
	ids := &identity.Sequence{}
	now := clock.System{}.Now()
	if err := application.Bootstrap(ctx, db, ids, cfg.DefaultTenant, now); err != nil {
		return err
	}
	coordinator := application.NewCoordinator(db, clock.System{}, ids)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	api := httpapi.New(coordinator, cfg.WebRoot, logger)
	server := &http.Server{Addr: cfg.Address, Handler: api.Handler(), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second}
	outboxWorker := worker.NewOutboxWorker(db, worker.LogPublisher{}, cfg.WorkerInterval, cfg.WorkerBatchSize)
	go outboxWorker.Run(ctx)
	serveErr := make(chan error, 1)
	go func() { logger.Info("windops listening", "address", cfg.Address); serveErr <- server.ListenAndServe() }()
	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
	}
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}
	outboxWorker.Wait()
	return nil
}
