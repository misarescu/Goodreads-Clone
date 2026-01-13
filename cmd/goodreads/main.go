// Package main provides the entry point for the Goodreads clone server application.
package main

import (
	"fmt"
	"log/slog"
	"os"

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

	slog.Info("Server starting", "address", addr)
	if err := srv.ListenAndServe(addr); err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}
