package storage

import (
	"context"
	"database/sql"
	"fmt"

	"goodreads-clone.com/internal/db"
)

// Storage provides access to all repository implementations and manages the database connection.
type Storage struct {
	queries *db.Queries
	db      *sql.DB

	Users           UserRepository
	Books           BookRepository
	Shelves         ShelfRepository
	ShelfBooks      ShelfBookRepository
	Challenges      ChallengeRepository
	Reviews         ReviewRepository
	Recommendations RecommendationRepository
}

// New creates a new Storage instance with the given database connection.
func New(sqlDB *sql.DB) *Storage {
	queries := db.New(sqlDB)
	s := &Storage{
		queries: queries,
		db:      sqlDB,
	}
	s.Users = &userRepository{queries: queries}
	s.Books = &bookRepository{queries: queries, db: sqlDB}
	s.Shelves = &shelfRepository{queries: queries}
	s.ShelfBooks = &shelfBookRepository{queries: queries}
	s.Challenges = &challengeRepository{queries: queries}
	s.Reviews = &reviewRepository{queries: queries}
	s.Recommendations = &recommendationRepository{queries: queries}
	return s
}

// Close closes the database connection and releases all resources.
func (s *Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	return nil
}

// DB returns the underlying database connection.
func (s *Storage) DB() *sql.DB {
	return s.db
}

// Ping reports whether the database answers.
func (s *Storage) Ping(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("pinging database: %w", err)
	}
	return nil
}
