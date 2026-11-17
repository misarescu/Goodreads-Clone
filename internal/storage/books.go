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
	db      *sql.DB
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

// ListByTag returns one page of the books carrying tag, newest first.
func (r *bookRepository) ListByTag(ctx context.Context, tag string, limit, offset int64) ([]db.Book, error) {
	books, err := r.queries.ListBooksByTag(ctx, db.ListBooksByTagParams{Tag: tag, Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list books by tag query failed: %w", err)
	}
	return books, nil
}

// SetTags replaces all of a book's tags with tags (already validated and unique) in one transaction.
func (r *bookRepository) SetTags(ctx context.Context, bookID int64, tags []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin set tags transaction failed: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	q := r.queries.WithTx(tx)
	if err := q.DeleteBookTags(ctx, bookID); err != nil {
		return fmt.Errorf("delete book tags query failed: %w", err)
	}
	for _, tag := range tags {
		if err := q.InsertBookTag(ctx, db.InsertBookTagParams{BookID: bookID, Tag: tag}); err != nil {
			return fmt.Errorf("insert book tag query failed: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit set tags failed: %w", err)
	}
	return nil
}

// TagsFor returns the sorted tags of each of the given books; books without tags are absent.
func (r *bookRepository) TagsFor(ctx context.Context, bookIDs []int64) (map[int64][]string, error) {
	out := map[int64][]string{}
	if len(bookIDs) == 0 {
		return out, nil
	}
	rows, err := r.queries.ListBookTags(ctx, bookIDs)
	if err != nil {
		return nil, fmt.Errorf("list book tags query failed: %w", err)
	}
	for _, row := range rows {
		out[row.BookID] = append(out[row.BookID], row.Tag)
	}
	return out, nil
}

// ListTagCounts returns one page of the tags in use that start with prefix (empty means all),
// ordered by number of books (most first) then tag, and the total number of matching tags.
func (r *bookRepository) ListTagCounts(
	ctx context.Context, prefix string, limit, offset int64,
) ([]TagCount, int64, error) {
	total, err := r.queries.CountTags(ctx, prefix)
	if err != nil {
		return nil, 0, fmt.Errorf("count tags query failed: %w", err)
	}
	rows, err := r.queries.ListTagCounts(ctx, db.ListTagCountsParams{Prefix: prefix, Limit: limit, Offset: offset})
	if err != nil {
		return nil, 0, fmt.Errorf("list tag counts query failed: %w", err)
	}
	out := make([]TagCount, len(rows))
	for i, row := range rows {
		out[i] = TagCount{Tag: row.Tag, Books: row.Books}
	}
	return out, total, nil
}
