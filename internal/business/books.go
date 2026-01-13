// Package business implements the core business logic layer for the Goodreads clone application.
package business

import (
	"context"
	"fmt"

	"goodreads-clone.com/internal/db"
)

// CreateBook creates a new book in the catalog.
func (b *Business) CreateBook(ctx context.Context, params db.CreateBookParams) (*db.Book, error) {
	book, err := b.Storage.Books.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create book: %w", err)
	}
	return book, nil
}

// GetBookByID retrieves a book by its ID.
func (b *Business) GetBookByID(ctx context.Context, id int64) (*db.Book, error) {
	book, err := b.Storage.Books.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get book by ID: %w", err)
	}
	return book, nil
}

// GetBookByISBN retrieves a book by its ISBN.
func (b *Business) GetBookByISBN(ctx context.Context, isbn string) (*db.Book, error) {
	book, err := b.Storage.Books.GetByISBN(ctx, isbn)
	if err != nil {
		return nil, fmt.Errorf("failed to get book by ISBN: %w", err)
	}
	return book, nil
}

// UpdateBook updates a book's information.
func (b *Business) UpdateBook(ctx context.Context, params db.UpdateBookParams) (*db.Book, error) {
	book, err := b.Storage.Books.Update(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update book: %w", err)
	}
	return book, nil
}

// DeleteBook deletes a book by its ID.
func (b *Business) DeleteBook(ctx context.Context, id int64) error {
	if err := b.Storage.Books.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete book: %w", err)
	}
	return nil
}

// ListBooks retrieves a paginated list of books.
func (b *Business) ListBooks(ctx context.Context, limit, offset int64) ([]db.Book, error) {
	books, err := b.Storage.Books.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list books: %w", err)
	}
	return books, nil
}

// SearchBooksByTitle searches for books by title.
func (b *Business) SearchBooksByTitle(ctx context.Context, query string, limit int64) ([]db.Book, error) {
	books, err := b.Storage.Books.SearchByTitle(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search books by title: %w", err)
	}
	return books, nil
}
