package business

import (
	"database/sql"
	"fmt"

	"goodreads-clone.com/internal/storage"
)

// Business coordinates business logic operations and manages the storage layer.
type Business struct {
	Storage *storage.Storage
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

	return &Business{
		Storage: storage.New(db),
	}, nil
}

// Close releases all resources held by the Business instance.
func (b *Business) Close() error {
	return b.Storage.Close()
}
