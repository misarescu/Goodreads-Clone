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
	db      db.DBTX
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

func (r *bookRepository) Count(ctx context.Context) (int64, error) {
	n, err := r.queries.CountBooks(ctx)
	if err != nil {
		return 0, fmt.Errorf("count books query failed: %w", err)
	}
	return n, nil
}

// Search runs a full-text query (an FTS5 MATCH expression, see business.ParseSearchQuery)
// and returns one page of books ranked by bm25 (title weighted above author), plus the total
// number of matches. The SQL is hand-written because sqlc cannot parse FTS5's MATCH.
func (r *bookRepository) Search(ctx context.Context, params SearchParams) ([]db.Book, int64, error) {
	const from = `FROM books b
JOIN books_fts(?1) ON books_fts.rowid = b.id
WHERE (?2 = '' OR lower(b.language) = lower(?2))`

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) "+from, params.Match, params.Language).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count search books query failed: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		"SELECT b.id, b.title, b.author, b.page_count, b.isbn, b.release_date, b.language, b.created_at, b.rating, "+
			"b.cover_url "+
			from+" ORDER BY bm25(books_fts, 10.0, 1.0), b.id LIMIT ?3 OFFSET ?4",
		params.Match, params.Language, params.Limit, params.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("search books query failed: %w", err)
	}
	defer func() { _ = rows.Close() }()

	books := []db.Book{}
	for rows.Next() {
		var b db.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.PageCount, &b.Isbn,
			&b.ReleaseDate, &b.Language, &b.CreatedAt, &b.Rating, &b.CoverUrl); err != nil {
			return nil, 0, fmt.Errorf("scan searched book failed: %w", err)
		}
		books = append(books, b)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("search books rows failed: %w", err)
	}
	return books, total, nil
}
