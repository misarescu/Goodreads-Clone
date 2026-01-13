package business

import (
	"context"

	"goodreads-clone.com/internal/db"
)

// CreateBook creates a new book in the catalog.
func (b *Business) CreateBook(ctx context.Context, params db.CreateBookParams) (*db.Book, error) {
	return b.Storage.Books.Create(ctx, params)
}

// GetBookByID retrieves a book by its ID.
func (b *Business) GetBookByID(ctx context.Context, id int64) (*db.Book, error) {
	return b.Storage.Books.GetByID(ctx, id)
}

// GetBookByISBN retrieves a book by its ISBN.
func (b *Business) GetBookByISBN(ctx context.Context, isbn string) (*db.Book, error) {
	return b.Storage.Books.GetByISBN(ctx, isbn)
}

// UpdateBook updates a book's information.
func (b *Business) UpdateBook(ctx context.Context, params db.UpdateBookParams) (*db.Book, error) {
	return b.Storage.Books.Update(ctx, params)
}

// DeleteBook deletes a book by its ID.
func (b *Business) DeleteBook(ctx context.Context, id int64) error {
	return b.Storage.Books.Delete(ctx, id)
}

// ListBooks retrieves a paginated list of books.
func (b *Business) ListBooks(ctx context.Context, limit, offset int64) ([]db.Book, error) {
	return b.Storage.Books.List(ctx, limit, offset)
}

// SearchBooksByTitle searches for books by title.
func (b *Business) SearchBooksByTitle(ctx context.Context, query string, limit int64) ([]db.Book, error) {
	return b.Storage.Books.SearchByTitle(ctx, query, limit)
}
