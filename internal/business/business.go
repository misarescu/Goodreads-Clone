package business

import (
	"database/sql"
	"fmt"

	"goodreads-clone.com/internal/storage"
)

type Business struct {
	Storage *storage.Storage
}

type Config struct {
	DatabasePath string
}

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

func (b *Business) Close() error {
	return b.Storage.Close()
}
