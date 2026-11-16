package business

import (
	"context"
	"database/sql"
	"errors"
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

// GetUserReadingStats returns reading stats for a user. It returns an error
// wrapping sql.ErrNoRows when the user does not exist.
func (b *Business) GetUserReadingStats(ctx context.Context, userID int64) (*db.GetUserReadingStatsRow, error) {
	if _, err := b.Storage.Users.GetByID(ctx, userID); err != nil {
		return nil, fmt.Errorf("failed to get user for reading stats: %w", err)
	}
	stats, err := b.Storage.ShelfBooks.GetUserReadingStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user reading stats: %w", err)
	}
	return stats, nil
}

// ErrInvalidProgress is returned when a progress update is not valid for the book.
var ErrInvalidProgress = errors.New("invalid reading progress")

func invalidProgress(format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{ErrInvalidProgress}, args...)...)
}

// computeProgress returns the pages read and percentage for a progress update.
func computeProgress(
	book *db.Book, current int64, pagesRead, progressPercent *int64,
) (pages, percent int64, err error) {
	if book.PageCount.Valid && book.PageCount.Int64 > 0 {
		pages, percent, err = progressFromPages(book.PageCount.Int64, pagesRead, progressPercent)
		return pages, percent, err
	}
	if progressPercent == nil || pagesRead != nil {
		return 0, 0, invalidProgress("progress_percent is required for a book without a page count")
	}
	if *progressPercent < 0 || *progressPercent > 100 {
		return 0, 0, invalidProgress("progress_percent must be between 0 and 100")
	}
	return current, *progressPercent, nil
}

func progressFromPages(total int64, pagesRead, progressPercent *int64) (pages, percent int64, err error) {
	if pagesRead == nil || progressPercent != nil {
		return 0, 0, invalidProgress("pages_read is required for a book with a page count")
	}
	if *pagesRead < 0 || *pagesRead > total {
		return 0, 0, invalidProgress("pages_read must be between 0 and %d", total)
	}
	return *pagesRead, min(*pagesRead*100/total, 100), nil
}

// RecordReadingProgress records progress for a book on a shelf. For a book with
// a page count, pagesRead is required and the percentage is computed (rounded
// down, capped at 100). For a book without one, progressPercent is required.
// It returns an error wrapping sql.ErrNoRows when the book is not on the shelf
// and one wrapping ErrInvalidProgress when the input is not valid.
func (b *Business) RecordReadingProgress(
	ctx context.Context,
	shelfID, bookID int64,
	pagesRead, progressPercent *int64,
) (*db.ShelfBook, error) {
	shelfBook, err := b.GetShelfBook(ctx, shelfID, bookID)
	if err != nil {
		return nil, err
	}
	book, err := b.Storage.Books.GetByID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get book for progress: %w", err)
	}
	pages, percent, err := computeProgress(book, shelfBook.PagesRead, pagesRead, progressPercent)
	if err != nil {
		return nil, err
	}
	return b.UpdateReadingProgress(ctx, shelfBook.ID, pages, percent)
}

// Errors returned by MoveBookToShelf.
var (
	ErrSameShelf          = errors.New("target shelf is the same as the source shelf")
	ErrDifferentUsers     = errors.New("shelves belong to different users")
	ErrBookAlreadyOnShelf = errors.New("book is already on the target shelf")
)

// MoveBookToShelf moves a book from one shelf to another shelf of the same
// user, keeping its reading progress. Missing shelves or a book that is not on
// the source shelf wrap sql.ErrNoRows.
func (b *Business) MoveBookToShelf(
	ctx context.Context,
	fromShelfID, bookID, toShelfID int64,
) (*db.ListBooksOnShelfRow, error) {
	from, err := b.Storage.Shelves.GetByID(ctx, fromShelfID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source shelf: %w", err)
	}
	to, err := b.Storage.Shelves.GetByID(ctx, toShelfID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target shelf: %w", err)
	}
	if from.ID == to.ID {
		return nil, ErrSameShelf
	}
	if from.UserID != to.UserID {
		return nil, ErrDifferentUsers
	}
	if _, err := b.Storage.ShelfBooks.Get(ctx, fromShelfID, bookID); err != nil {
		return nil, fmt.Errorf("failed to get shelf book: %w", err)
	}
	if _, err := b.Storage.ShelfBooks.Get(ctx, toShelfID, bookID); err == nil {
		return nil, ErrBookAlreadyOnShelf
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check target shelf: %w", err)
	}
	moved, err := b.Storage.ShelfBooks.Move(ctx, fromShelfID, bookID, toShelfID)
	if err != nil {
		return nil, fmt.Errorf("failed to move book: %w", err)
	}
	return moved, nil
}

// MinStatsYear is the earliest year GetMonthlyReadingStats accepts; the latest is next year.
const MinStatsYear = 1900

// ErrInvalidYear is returned when the requested stats year is out of range.
var ErrInvalidYear = errors.New("invalid year")

// MonthlyReading is one month of a user's reading stats.
type MonthlyReading struct {
	Month         int   `json:"month"`
	BooksFinished int64 `json:"books_finished"`
	PagesFinished int64 `json:"pages_finished"`
}

// GetMonthlyReadingStats returns twelve entries, January to December, with the
// books and pages the user finished each month of the year; a nil year means
// the current year. It returns ErrInvalidYear for a year before MinStatsYear or
// after next year, and an error wrapping sql.ErrNoRows for an unknown user.
func (b *Business) GetMonthlyReadingStats(ctx context.Context, userID int64, year *int64) ([]MonthlyReading, error) {
	current := int64(b.now().UTC().Year())
	y := current
	if year != nil {
		y = *year
	}
	if y < MinStatsYear || y > current+1 {
		return nil, fmt.Errorf("%w: year must be between %d and %d", ErrInvalidYear, MinStatsYear, current+1)
	}
	if _, err := b.Storage.Users.GetByID(ctx, userID); err != nil {
		return nil, fmt.Errorf("failed to get user for monthly stats: %w", err)
	}
	rows, err := b.Storage.ShelfBooks.ListMonthlyFinished(ctx, userID, y)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly reading stats: %w", err)
	}
	out := make([]MonthlyReading, 12)
	for i := range out {
		out[i].Month = i + 1
	}
	for _, r := range rows {
		if r.Month >= 1 && r.Month <= 12 {
			out[r.Month-1].BooksFinished = r.BooksFinished
			out[r.Month-1].PagesFinished = r.PagesFinished
		}
	}
	return out, nil
}
