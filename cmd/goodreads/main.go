// Package main provides the entry point for the Goodreads clone server application.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "modernc.org/sqlite"

	"goodreads-clone.com/internal/business"
	"goodreads-clone.com/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Application error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "goodreads.db"
	}

	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	shutdownTimeout := 20 * time.Second
	if v := os.Getenv("SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			return fmt.Errorf("invalid SHUTDOWN_TIMEOUT %q: want a positive Go duration such as 20s", v)
		}
		shutdownTimeout = d
	}

	b, err := business.New(business.Config{
		DatabasePath: dbPath,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize business layer: %w", err)
	}
	defer func() { _ = b.Close() }()

	srv, err := server.New(b)
	if err != nil {
		return fmt.Errorf("failed to initialize server: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()

	slog.Info("Server starting", "address", addr)
	if err := srv.ListenAndServe(ctx, addr, shutdownTimeout); err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	slog.Info("Server stopped")
	return nil
}
