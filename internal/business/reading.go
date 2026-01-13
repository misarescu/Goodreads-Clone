package business

import (
	"context"

	"goodreads-clone.com/internal/db"
)

// AddBookToShelf adds a book to a shelf with initial progress.
func (b *Business) AddBookToShelf(ctx context.Context, shelfID, bookID, pagesRead, progressPercent int64) (*db.ShelfBook, error) {
	return b.Storage.ShelfBooks.Add(ctx, shelfID, bookID, pagesRead, progressPercent)
}

// GetShelfBook retrieves a book's entry on a specific shelf.
func (b *Business) GetShelfBook(ctx context.Context, shelfID, bookID int64) (*db.ShelfBook, error) {
	return b.Storage.ShelfBooks.Get(ctx, shelfID, bookID)
}

// UpdateReadingProgress updates the reading progress for a book on a shelf.
func (b *Business) UpdateReadingProgress(ctx context.Context, id, pagesRead, progressPercent int64) (*db.ShelfBook, error) {
	return b.Storage.ShelfBooks.UpdateProgress(ctx, id, pagesRead, progressPercent)
}

// RemoveBookFromShelf removes a book from a shelf.
func (b *Business) RemoveBookFromShelf(ctx context.Context, shelfID, bookID int64) error {
	return b.Storage.ShelfBooks.Remove(ctx, shelfID, bookID)
}

// ListBooksOnShelf lists all books on a specific shelf with their details.
func (b *Business) ListBooksOnShelf(ctx context.Context, shelfID int64) ([]db.ListBooksOnShelfRow, error) {
	return b.Storage.ShelfBooks.ListBooksOnShelf(ctx, shelfID)
}

// GetUserBookShelves returns all shelves that contain a specific book for a user.
func (b *Business) GetUserBookShelves(ctx context.Context, userID, bookID int64) ([]db.Shelf, error) {
	return b.Storage.ShelfBooks.GetUserBookShelves(ctx, userID, bookID)
}
