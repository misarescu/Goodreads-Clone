package storage

import (
	"context"
	"database/sql"

	"goodreads-clone.com/internal/db"
)

type bookRepository struct {
	queries *db.Queries
}

func (r *bookRepository) Create(ctx context.Context, params db.CreateBookParams) (*db.Book, error) {
	book, err := r.queries.CreateBook(ctx, params)
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func (r *bookRepository) GetByID(ctx context.Context, id int64) (*db.Book, error) {
	book, err := r.queries.GetBookByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func (r *bookRepository) GetByISBN(ctx context.Context, isbn string) (*db.Book, error) {
	book, err := r.queries.GetBookByISBN(ctx, sql.NullString{String: isbn, Valid: true})
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func (r *bookRepository) Update(ctx context.Context, params db.UpdateBookParams) (*db.Book, error) {
	book, err := r.queries.UpdateBook(ctx, params)
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func (r *bookRepository) Delete(ctx context.Context, id int64) error {
	return r.queries.DeleteBook(ctx, id)
}

func (r *bookRepository) List(ctx context.Context, limit, offset int64) ([]db.Book, error) {
	return r.queries.ListBooks(ctx, db.ListBooksParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *bookRepository) SearchByTitle(ctx context.Context, query string, limit int64) ([]db.Book, error) {
	return r.queries.SearchBooksByTitle(ctx, db.SearchBooksByTitleParams{
		Title: "%" + query + "%",
		Limit: limit,
	})
}
