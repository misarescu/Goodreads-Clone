// Package storage implements the data access layer for the Goodreads clone application.
package storage

import (
	"context"
	"database/sql"
	"fmt"

	"goodreads-clone.com/internal/db"
)

type bookRepository struct {
	queries *db.Queries
}

func (r *bookRepository) Create(ctx context.Context, params db.CreateBookParams) (*db.Book, error) {
	book, err := r.queries.CreateBook(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create book query failed: %w", err)
	}
	return &book, nil
}

func (r *bookRepository) GetByID(ctx context.Context, id int64) (*db.Book, error) {
	book, err := r.queries.GetBookByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get book by ID query failed: %w", err)
	}
	return &book, nil
}

func (r *bookRepository) GetByISBN(ctx context.Context, isbn string) (*db.Book, error) {
	book, err := r.queries.GetBookByISBN(ctx, sql.NullString{String: isbn, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("get book by ISBN query failed: %w", err)
	}
	return &book, nil
}

func (r *bookRepository) Update(ctx context.Context, params db.UpdateBookParams) (*db.Book, error) {
	book, err := r.queries.UpdateBook(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update book query failed: %w", err)
	}
	return &book, nil
}

func (r *bookRepository) Delete(ctx context.Context, id int64) error {
	if err := r.queries.DeleteBook(ctx, id); err != nil {
		return fmt.Errorf("delete book query failed: %w", err)
	}
	return nil
}

func (r *bookRepository) List(ctx context.Context, limit, offset int64) ([]db.Book, error) {
	books, err := r.queries.ListBooks(ctx, db.ListBooksParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list books query failed: %w", err)
	}
	return books, nil
}

func (r *bookRepository) SearchByTitle(ctx context.Context, query string, limit int64) ([]db.Book, error) {
	books, err := r.queries.SearchBooksByTitle(ctx, db.SearchBooksByTitleParams{
		Title: "%" + query + "%",
		Limit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("search books by title query failed: %w", err)
	}
	return books, nil
}
