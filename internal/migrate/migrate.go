// Package migrate applies SQL migrations to the database and records them in schema_migrations.
package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

// legacyBaseline is the last migration applied by `make migrate` before
// schema_migrations existed.
const legacyBaseline = "006_books_fts.sql"

// Apply runs every *.sql file in dir of fsys that is not yet recorded in
// schema_migrations, in file name order, each in its own transaction.
// A database that has the schema but no schema_migrations table (created by
// `make migrate`) gets the files up to 006 recorded without running them.
func Apply(ctx context.Context, db *sql.DB, fsys fs.FS, dir string) error {
	files, err := fs.Glob(fsys, dir+"/*.sql")
	if err != nil {
		return fmt.Errorf("listing migrations: %w", err)
	}
	sort.Strings(files)

	legacy, err := isLegacy(ctx, db)
	if err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		name TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("creating schema_migrations: %w", err)
	}
	if legacy {
		if err := recordBaseline(ctx, db, files); err != nil {
			return err
		}
	}

	for _, f := range files {
		if err := applyFile(ctx, db, fsys, f); err != nil {
			return err
		}
	}
	return nil
}

// isLegacy reports whether the schema exists without schema_migrations.
func isLegacy(ctx context.Context, db *sql.DB) (bool, error) {
	tracked, err := tableExists(ctx, db, "schema_migrations")
	if err != nil || tracked {
		return false, err
	}
	return tableExists(ctx, db, "users")
}

func recordBaseline(ctx context.Context, db *sql.DB, files []string) error {
	for _, f := range files {
		name := baseName(f)
		if name > legacyBaseline {
			break
		}
		if _, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO schema_migrations (name) VALUES (?)`, name); err != nil {
			return fmt.Errorf("recording existing migration %s: %w", name, err)
		}
	}
	return nil
}

func applyFile(ctx context.Context, db *sql.DB, fsys fs.FS, f string) error {
	name := baseName(f)
	var n int
	q := `SELECT COUNT(*) FROM schema_migrations WHERE name = ?`
	if err := db.QueryRowContext(ctx, q, name).Scan(&n); err != nil {
		return fmt.Errorf("checking migration %s: %w", name, err)
	}
	if n > 0 {
		return nil
	}
	body, err := fs.ReadFile(fsys, f)
	if err != nil {
		return fmt.Errorf("reading migration %s: %w", name, err)
	}
	if err := applyOne(ctx, db, name, string(body)); err != nil {
		return fmt.Errorf("migration %s: %w", name, err)
	}
	return nil
}

func applyOne(ctx context.Context, db *sql.DB, name, body string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("executing: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, body); err != nil {
		return fmt.Errorf("executing: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (name) VALUES (?)`, name); err != nil {
		return fmt.Errorf("executing: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing: %w", err)
	}
	return nil
}

func tableExists(ctx context.Context, db *sql.DB, name string) (bool, error) {
	var n int
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, name).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("checking for table %s: %w", name, err)
	}
	return n > 0, nil
}

func baseName(path string) string {
	return path[strings.LastIndex(path, "/")+1:]
}
