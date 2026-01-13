package storage

import (
	"database/sql"

	"goodreads-clone.com/internal/db"
)

type Storage struct {
	queries *db.Queries
	db      *sql.DB

	Users      UserRepository
	Books      BookRepository
	Shelves    ShelfRepository
	ShelfBooks ShelfBookRepository
}

func New(sqlDB *sql.DB) *Storage {
	queries := db.New(sqlDB)
	s := &Storage{
		queries: queries,
		db:      sqlDB,
	}
	s.Users = &userRepository{queries: queries}
	s.Books = &bookRepository{queries: queries}
	s.Shelves = &shelfRepository{queries: queries}
	s.ShelfBooks = &shelfBookRepository{queries: queries}
	return s
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) DB() *sql.DB {
	return s.db
}
