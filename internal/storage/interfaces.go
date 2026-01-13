package storage

import (
	"context"

	"goodreads-clone.com/internal/db"
)

// UserRepository defines operations for managing user data.
type UserRepository interface {
	Create(ctx context.Context, email, username, passwordHash string) (*db.User, error)
	GetByID(ctx context.Context, id int64) (*db.User, error)
	GetByEmail(ctx context.Context, email string) (*db.User, error)
	GetByUsername(ctx context.Context, username string) (*db.User, error)
	Update(ctx context.Context, id int64, email, username string) (*db.User, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int64) ([]db.User, error)
}

// BookRepository defines operations for managing book data.
type BookRepository interface {
	Create(ctx context.Context, params db.CreateBookParams) (*db.Book, error)
	GetByID(ctx context.Context, id int64) (*db.Book, error)
	GetByISBN(ctx context.Context, isbn string) (*db.Book, error)
	Update(ctx context.Context, params db.UpdateBookParams) (*db.Book, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int64) ([]db.Book, error)
	SearchByTitle(ctx context.Context, query string, limit int64) ([]db.Book, error)
}

// ShelfRepository defines operations for managing user shelves.
type ShelfRepository interface {
	Create(ctx context.Context, userID int64, name string, isDefault bool) (*db.Shelf, error)
	GetByID(ctx context.Context, id int64) (*db.Shelf, error)
	GetByUserAndName(ctx context.Context, userID int64, name string) (*db.Shelf, error)
	ListByUser(ctx context.Context, userID int64) ([]db.Shelf, error)
	Update(ctx context.Context, id int64, name string) (*db.Shelf, error)
	Delete(ctx context.Context, id int64) error
	GetDefaultsByUser(ctx context.Context, userID int64) ([]db.Shelf, error)
}

// ShelfBookRepository defines operations for managing books on shelves and reading progress.
type ShelfBookRepository interface {
	Add(ctx context.Context, shelfID, bookID, pagesRead, progressPercent int64) (*db.ShelfBook, error)
	Get(ctx context.Context, shelfID, bookID int64) (*db.ShelfBook, error)
	UpdateProgress(ctx context.Context, id, pagesRead, progressPercent int64) (*db.ShelfBook, error)
	Remove(ctx context.Context, shelfID, bookID int64) error
	ListBooksOnShelf(ctx context.Context, shelfID int64) ([]db.ListBooksOnShelfRow, error)
	GetUserBookShelves(ctx context.Context, userID, bookID int64) ([]db.Shelf, error)
}
