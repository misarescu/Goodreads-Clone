package main

import (
	"database/sql"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	dbmigrations "goodreads-clone.com/db"
)

func open(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func applied(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.Query(`SELECT name || '@' || applied_at FROM schema_migrations ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		out = append(out, s)
	}
	return out
}

func allFiles(t *testing.T) []string {
	t.Helper()
	files, err := fs.Glob(dbmigrations.FS, "migrations/*.sql")
	if err != nil || len(files) == 0 {
		t.Fatalf("no migrations: %v", err)
	}
	sort.Strings(files)
	return files
}

func TestRunFreshDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fresh.db")
	if err := run(path); err != nil {
		t.Fatal(err)
	}
	if got, want := len(applied(t, open(t, path))), len(allFiles(t)); got != want {
		t.Fatalf("recorded %d migrations, want %d", got, want)
	}
}

func TestRunTwiceAppliesNothing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "twice.db")
	if err := run(path); err != nil {
		t.Fatal(err)
	}
	first := applied(t, open(t, path))
	if err := run(path); err != nil {
		t.Fatalf("second run: %v", err)
	}
	second := applied(t, open(t, path))
	if strings.Join(first, ",") != strings.Join(second, ",") {
		t.Fatalf("second run changed schema_migrations:\n%v\n%v", first, second)
	}
}

// TestRunLegacyDatabase simulates the old `make migrate` loop: migrations up
// to 006_books_fts.sql applied directly, no schema_migrations table.
func TestRunLegacyDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	db := open(t, path)
	for _, f := range allFiles(t) {
		if filepath.Base(f) > "006_books_fts.sql" {
			break
		}
		body, err := fs.ReadFile(dbmigrations.FS, f)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(string(body)); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
	}
	if err := run(path); err != nil {
		t.Fatal(err)
	}
	if got, want := len(applied(t, db)), len(allFiles(t)); got != want {
		t.Fatalf("recorded %d migrations, want %d", got, want)
	}
}
