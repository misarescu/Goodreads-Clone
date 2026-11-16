package migrate_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"testing/fstest"

	_ "modernc.org/sqlite"

	dbmigrations "goodreads-clone.com/db"
	"goodreads-clone.com/internal/migrate"
)

func open(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func count(t *testing.T, db *sql.DB, q string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(q).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestApplyFreshDatabase(t *testing.T) {
	db := open(t)
	ctx := context.Background()
	for i := 0; i < 2; i++ { // second run is a no-op
		if err := migrate.Apply(ctx, db, dbmigrations.FS, "migrations"); err != nil {
			t.Fatal(err)
		}
	}
	if n := count(t, db, "SELECT COUNT(*) FROM schema_migrations"); n != 7 {
		t.Fatalf("recorded = %d, want 7", n)
	}
	if _, err := db.Exec("INSERT INTO books (title, author) VALUES ('a','b')"); err != nil {
		t.Fatal(err)
	}
}

func TestApplyLegacyDatabase(t *testing.T) {
	db := open(t)
	ctx := context.Background()
	// Simulate `make migrate` before schema_migrations existed: files up to 006 run
	// directly, no schema_migrations; 007 is applied by Apply.
	entries, _ := dbmigrations.FS.ReadDir("migrations")
	for _, e := range entries {
		if e.Name() > "006_books_fts.sql" {
			continue
		}
		b, _ := dbmigrations.FS.ReadFile("migrations/" + e.Name())
		if _, err := db.Exec(string(b)); err != nil {
			t.Fatal(err)
		}
	}
	// Re-running 002 would fail (duplicate column), so success proves it was skipped.
	if err := migrate.Apply(ctx, db, dbmigrations.FS, "migrations"); err != nil {
		t.Fatal(err)
	}
	if n := count(t, db, "SELECT COUNT(*) FROM schema_migrations"); n != 7 {
		t.Fatalf("recorded = %d, want 7", n)
	}
}

func TestApplyPicksUpNewFileAndReportsFailure(t *testing.T) {
	db := open(t)
	ctx := context.Background()
	fsys := fstest.MapFS{
		"m/001_a.sql": {Data: []byte("CREATE TABLE a (id INTEGER);")},
	}
	if err := migrate.Apply(ctx, db, fsys, "m"); err != nil {
		t.Fatal(err)
	}
	fsys["m/002_b.sql"] = &fstest.MapFile{Data: []byte("CREATE TABLE b (id INTEGER);")}
	if err := migrate.Apply(ctx, db, fsys, "m"); err != nil {
		t.Fatal(err)
	}
	if n := count(t, db, "SELECT COUNT(*) FROM b"); n != 0 {
		t.Fatal("table b missing")
	}

	fsys["m/003_bad.sql"] = &fstest.MapFile{Data: []byte("CREATE TABLE c (id INTEGER); SELECT * FROM nope;")}
	err := migrate.Apply(ctx, db, fsys, "m")
	if err == nil || !contains(err.Error(), "003_bad.sql") {
		t.Fatalf("err = %v, want it to name 003_bad.sql", err)
	}
	if n := count(t, db, "SELECT COUNT(*) FROM sqlite_master WHERE name = 'c'"); n != 0 {
		t.Fatal("failed migration was not rolled back")
	}
	if n := count(t, db, "SELECT COUNT(*) FROM schema_migrations"); n != 2 {
		t.Fatalf("recorded = %d, want 2", n)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
