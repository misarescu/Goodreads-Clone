package business

import (
	"context"
	"fmt"

	"goodreads-clone.com/internal/db"
)

// AddBookToShelf adds a book to a shelf with initial progress.
func (b *Business) AddBookToShelf(
	ctx context.Context,
	shelfID, bookID, pagesRead, progressPercent int64,
) (*db.ShelfBook, error) {
	shelfBook, err := b.Storage.ShelfBooks.Add(ctx, shelfID, bookID, pagesRead, progressPercent)
	if err != nil {
		return nil, fmt.Errorf("failed to add book to shelf: %w", err)
	}
	return shelfBook, nil
}

// GetShelfBook retrieves a book's entry on a specific shelf.
func (b *Business) GetShelfBook(ctx context.Context, shelfID, bookID int64) (*db.ShelfBook, error) {
	shelfBook, err := b.Storage.ShelfBooks.Get(ctx, shelfID, bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shelf book: %w", err)
	}
	return shelfBook, nil
}

// UpdateReadingProgress updates the reading progress for a book on a shelf.
func (b *Business) UpdateReadingProgress(
	ctx context.Context,
	id, pagesRead, progressPercent int64,
) (*db.ShelfBook, error) {
	shelfBook, err := b.Storage.ShelfBooks.UpdateProgress(ctx, id, pagesRead, progressPercent)
	if err != nil {
		return nil, fmt.Errorf("failed to update reading progress: %w", err)
	}
	return shelfBook, nil
}

// RemoveBookFromShelf removes a book from a shelf.
func (b *Business) RemoveBookFromShelf(ctx context.Context, shelfID, bookID int64) error {
	if err := b.Storage.ShelfBooks.Remove(ctx, shelfID, bookID); err != nil {
		return fmt.Errorf("failed to remove book from shelf: %w", err)
	}
	return nil
}

// ListBooksOnShelf lists all books on a specific shelf with their details.
func (b *Business) ListBooksOnShelf(ctx context.Context, shelfID int64) ([]db.ListBooksOnShelfRow, error) {
	books, err := b.Storage.ShelfBooks.ListBooksOnShelf(ctx, shelfID)
	if err != nil {
		return nil, fmt.Errorf("failed to list books on shelf: %w", err)
	}
	return books, nil
}

// GetUserBookShelves returns all shelves that contain a specific book for a user.
func (b *Business) GetUserBookShelves(ctx context.Context, userID, bookID int64) ([]db.Shelf, error) {
	shelves, err := b.Storage.ShelfBooks.GetUserBookShelves(ctx, userID, bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user book shelves: %w", err)
	}
	return shelves, nil
}
