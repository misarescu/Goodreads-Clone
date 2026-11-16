package business

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	dbmigrations "goodreads-clone.com/db"
	"goodreads-clone.com/internal/migrate"
	"goodreads-clone.com/internal/storage"
)

// Business coordinates business logic operations and manages the storage layer.
type Business struct {
	Storage *storage.Storage
	// Now returns the current time; nil means time.Now. Tests set it to fix the clock.
	Now func() time.Time
}

// Config holds configuration options for initializing the Business layer.
type Config struct {
	DatabasePath string
}

// New creates a new Business instance with the given configuration.
// It opens the database connection and initializes the storage layer.
func New(cfg Config) (*Business, error) {
	db, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Enable foreign keys for SQLite
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	if err := migrate.Apply(context.Background(), db, dbmigrations.FS, "migrations"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("applying migrations: %w", err)
	}

	return &Business{
		Storage: storage.New(db),
	}, nil
}

// Close releases all resources held by the Business instance.
func (b *Business) Close() error {
	if err := b.Storage.Close(); err != nil {
		return fmt.Errorf("closing storage: %w", err)
	}
	return nil
}

// Ready reports whether the database answers a ping.
func (b *Business) Ready(ctx context.Context) error {
	if err := b.Storage.Ping(ctx); err != nil {
		return fmt.Errorf("readiness check: %w", err)
	}
	return nil
}
