package storage

import (
	"context"
	"database/sql"
	"fmt"

	"goodreads-clone.com/internal/db"
)

type shelfBookRepository struct {
	db      *sql.DB
	queries *db.Queries
}

func (r *shelfBookRepository) Add(
	ctx context.Context,
	shelfID, bookID, pagesRead, progressPercent int64,
) (*db.ShelfBook, error) {
	shelfBook, err := r.queries.AddBookToShelf(ctx, db.AddBookToShelfParams{
		ShelfID:         shelfID,
		BookID:          bookID,
		PagesRead:       pagesRead,
		ProgressPercent: progressPercent,
	})
	if err != nil {
		return nil, fmt.Errorf("add book to shelf query failed: %w", err)
	}
	return &shelfBook, nil
}

func (r *shelfBookRepository) Get(ctx context.Context, shelfID, bookID int64) (*db.ShelfBook, error) {
	shelfBook, err := r.queries.GetShelfBook(ctx, db.GetShelfBookParams{
		ShelfID: shelfID,
		BookID:  bookID,
	})
	if err != nil {
		return nil, fmt.Errorf("get shelf book query failed: %w", err)
	}
	return &shelfBook, nil
}

func (r *shelfBookRepository) UpdateProgress(
	ctx context.Context,
	id, pagesRead, progressPercent int64,
) (*db.ShelfBook, error) {
	shelfBook, err := r.queries.UpdateShelfBookProgress(ctx, db.UpdateShelfBookProgressParams{
		ID:              id,
		PagesRead:       pagesRead,
		ProgressPercent: progressPercent,
	})
	if err != nil {
		return nil, fmt.Errorf("update shelf book progress query failed: %w", err)
	}
	return &shelfBook, nil
}

func (r *shelfBookRepository) Remove(ctx context.Context, shelfID, bookID int64) error {
	if err := r.queries.RemoveBookFromShelf(ctx, db.RemoveBookFromShelfParams{
		ShelfID: shelfID,
		BookID:  bookID,
	}); err != nil {
		return fmt.Errorf("remove book from shelf query failed: %w", err)
	}
	return nil
}

func (r *shelfBookRepository) ListBooksOnShelf(ctx context.Context, shelfID int64) ([]db.ListBooksOnShelfRow, error) {
	books, err := r.queries.ListBooksOnShelf(ctx, shelfID)
	if err != nil {
		return nil, fmt.Errorf("list books on shelf query failed: %w", err)
	}
	return books, nil
}

func (r *shelfBookRepository) GetUserBookShelves(ctx context.Context, userID, bookID int64) ([]db.Shelf, error) {
	shelves, err := r.queries.GetUserBookShelves(ctx, db.GetUserBookShelvesParams{
		UserID: userID,
		BookID: bookID,
	})
	if err != nil {
		return nil, fmt.Errorf("get user book shelves query failed: %w", err)
	}
	return shelves, nil
}

func (r *shelfBookRepository) GetUserReadingStats(
	ctx context.Context,
	userID int64,
) (*db.GetUserReadingStatsRow, error) {
	stats, err := r.queries.GetUserReadingStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user reading stats query failed: %w", err)
	}
	return &stats, nil
}

func (r *shelfBookRepository) Move(
	ctx context.Context,
	fromShelfID, bookID, toShelfID int64,
) (*db.ListBooksOnShelfRow, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin move transaction failed: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	q := r.queries.WithTx(tx)
	n, err := q.MoveShelfBook(ctx, db.MoveShelfBookParams{
		ToShelfID:   toShelfID,
		FromShelfID: fromShelfID,
		BookID:      bookID,
	})
	if err != nil {
		return nil, fmt.Errorf("move shelf book query failed: %w", err)
	}
	if n == 0 {
		return nil, fmt.Errorf("move shelf book: %w", sql.ErrNoRows)
	}
	row, err := q.GetBookOnShelf(ctx, db.GetBookOnShelfParams{ShelfID: toShelfID, BookID: bookID})
	if err != nil {
		return nil, fmt.Errorf("get moved shelf book query failed: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit move transaction failed: %w", err)
	}
	result := db.ListBooksOnShelfRow(row)
	return &result, nil
}

// ListForExport returns every book on the user's shelves, ordered by shelf name then book title.
func (r *shelfBookRepository) ListForExport(ctx context.Context, userID int64) ([]ShelfExportRow, error) {
	rows, err := r.queries.ListUserShelfBooksForExport(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list shelf books for export query failed: %w", err)
	}
	out := make([]ShelfExportRow, 0, len(rows))
	for _, row := range rows {
		e := ShelfExportRow{
			Shelf:           row.Shelf,
			Title:           row.Title,
			Author:          row.Author,
			ISBN:            row.Isbn.String,
			PagesRead:       row.PagesRead,
			ProgressPercent: row.ProgressPercent,
			AddedAt:         row.AddedAt.UTC(),
		}
		if row.FinishedAt != "" {
			at, err := parseFinishedAt(row.FinishedAt)
			if err != nil {
				return nil, fmt.Errorf("list shelf books for export: %w", err)
			}
			e.FinishedAt = &at
		}
		out = append(out, e)
	}
	return out, nil
}
