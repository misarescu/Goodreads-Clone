package storage

import (
	"context"

	"goodreads-clone.com/internal/db"
)

type shelfBookRepository struct {
	queries *db.Queries
}

func (r *shelfBookRepository) Add(ctx context.Context, shelfID, bookID, pagesRead, progressPercent int64) (*db.ShelfBook, error) {
	shelfBook, err := r.queries.AddBookToShelf(ctx, db.AddBookToShelfParams{
		ShelfID:         shelfID,
		BookID:          bookID,
		PagesRead:       pagesRead,
		ProgressPercent: progressPercent,
	})
	if err != nil {
		return nil, err
	}
	return &shelfBook, nil
}

func (r *shelfBookRepository) Get(ctx context.Context, shelfID, bookID int64) (*db.ShelfBook, error) {
	shelfBook, err := r.queries.GetShelfBook(ctx, db.GetShelfBookParams{
		ShelfID: shelfID,
		BookID:  bookID,
	})
	if err != nil {
		return nil, err
	}
	return &shelfBook, nil
}

func (r *shelfBookRepository) UpdateProgress(ctx context.Context, id, pagesRead, progressPercent int64) (*db.ShelfBook, error) {
	shelfBook, err := r.queries.UpdateShelfBookProgress(ctx, db.UpdateShelfBookProgressParams{
		ID:              id,
		PagesRead:       pagesRead,
		ProgressPercent: progressPercent,
	})
	if err != nil {
		return nil, err
	}
	return &shelfBook, nil
}

func (r *shelfBookRepository) Remove(ctx context.Context, shelfID, bookID int64) error {
	return r.queries.RemoveBookFromShelf(ctx, db.RemoveBookFromShelfParams{
		ShelfID: shelfID,
		BookID:  bookID,
	})
}

func (r *shelfBookRepository) ListBooksOnShelf(ctx context.Context, shelfID int64) ([]db.ListBooksOnShelfRow, error) {
	return r.queries.ListBooksOnShelf(ctx, shelfID)
}

func (r *shelfBookRepository) GetUserBookShelves(ctx context.Context, userID, bookID int64) ([]db.Shelf, error) {
	return r.queries.GetUserBookShelves(ctx, db.GetUserBookShelvesParams{
		UserID: userID,
		BookID: bookID,
	})
}
