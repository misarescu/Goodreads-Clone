package storage

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"goodreads-clone.com/internal/db"

	_ "modernc.org/sqlite"
)

// setupTestDB creates an in-memory SQLite database and applies migrations
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	// Enable foreign keys (critical for testing constraints)
	if _, err := database.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	// Load and execute schema
	schemaPath := filepath.Join("..", "..", "db", "migrations", "001_schema.sql")
	schemaSQL, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("failed to read schema file: %v", err)
	}

	if _, err := database.Exec(string(schemaSQL)); err != nil {
		t.Fatalf("failed to execute schema: %v", err)
	}

	// Register cleanup
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("failed to close database: %v", err)
		}
	})

	return database
}

// newTestStorage creates a new Storage instance with an in-memory database
func newTestStorage(t *testing.T) *Storage {
	t.Helper()
	database := setupTestDB(t)
	return New(database)
}

// createTestUser is a helper function to create a test user
func createTestUser(t *testing.T, s *Storage, email, username string) *db.User {
	t.Helper()
	user, err := s.Users.Create(context.Background(), email, username, "hashed_password")
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

// createTestBook is a helper function to create a test book
func createTestBook(t *testing.T, s *Storage, title, author string) *db.Book {
	t.Helper()
	book, err := s.Books.Create(context.Background(), db.CreateBookParams{
		Title:  title,
		Author: author,
	})
	if err != nil {
		t.Fatalf("failed to create test book: %v", err)
	}
	return book
}

// createTestShelf is a helper function to create a test shelf
func createTestShelf(t *testing.T, s *Storage, userID int64, name string, isDefault bool) *db.Shelf {
	t.Helper()
	shelf, err := s.Shelves.Create(context.Background(), userID, name, isDefault)
	if err != nil {
		t.Fatalf("failed to create test shelf: %v", err)
	}
	return shelf
}
