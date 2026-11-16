// Command migrate applies the embedded database migrations and exits.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "modernc.org/sqlite"

	dbmigrations "goodreads-clone.com/db"
	"goodreads-clone.com/internal/migrate"
)

func main() {
	if err := run(os.Getenv("DATABASE_PATH")); err != nil {
		slog.Error("Migration failed", "error", err)
		os.Exit(1)
	}
}

// run applies pending migrations to the database at dbPath (default
// goodreads.db, like the server).
func run(dbPath string) error {
	if dbPath == "" {
		dbPath = "goodreads.db"
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("enabling foreign keys: %w", err)
	}
	if err := migrate.Apply(context.Background(), db, dbmigrations.FS, "migrations"); err != nil {
		return fmt.Errorf("applying migrations: %w", err)
	}
	slog.Info("Migrations complete", "database", dbPath)
	return nil
}
