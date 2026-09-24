package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aisuomii/shopik-catalog-service/internal/config"
	"github.com/aisuomii/shopik-catalog-service/internal/repository"
)

func main() {
	if err := run(); err != nil {
		slog.Error("service stopped", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := repository.Open(ctx, cfg.Postgres)
	if err != nil {
		return err
	}
	defer db.Close()

	slog.Info("connected to postgres",
		slog.String("env", cfg.App.Env),
		slog.String("host", cfg.Postgres.Host),
		slog.String("database", cfg.Postgres.Database),
	)

	// TODO: wire handler -> service -> repository and serve here.
	<-ctx.Done()

	return nil
}
