package storage

import (
	"context"
	"fmt"

	"goodreads-clone.com/internal/db"
)

type shelfBookRepository struct {
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
