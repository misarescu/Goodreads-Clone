package storage

import (
	"context"

	"goodreads-clone.com/internal/db"
)

// UserRepository defines operations for managing user data.
type UserRepository interface {
	Create(ctx context.Context, email, username, passwordHash string) (*db.User, error)
	GetByID(ctx context.Context, id int64) (*db.User, error)
	GetByEmail(ctx context.Context, email string) (*db.User, error)
	GetByUsername(ctx context.Context, username string) (*db.User, error)
	Update(ctx context.Context, id int64, email, username string) (*db.User, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int64) ([]db.User, error)
}

// SearchParams is a full-text book search: an FTS5 MATCH expression, an optional
// language filter (empty means any) and a page.
type SearchParams struct {
	Match    string
	Language string
	Limit    int64
	Offset   int64
}

// BookRepository defines operations for managing book data.
type BookRepository interface {
	Create(ctx context.Context, params db.CreateBookParams) (*db.Book, error)
	GetByID(ctx context.Context, id int64) (*db.Book, error)
	GetByISBN(ctx context.Context, isbn string) (*db.Book, error)
	Update(ctx context.Context, params db.UpdateBookParams) (*db.Book, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int64) ([]db.Book, error)
	Search(ctx context.Context, params SearchParams) ([]db.Book, int64, error)
	Count(ctx context.Context) (int64, error)
}

// ShelfRepository defines operations for managing user shelves.
type ShelfRepository interface {
	Create(ctx context.Context, userID int64, name string, isDefault bool) (*db.Shelf, error)
	GetByID(ctx context.Context, id int64) (*db.Shelf, error)
	GetByUserAndName(ctx context.Context, userID int64, name string) (*db.Shelf, error)
	ListByUser(ctx context.Context, userID int64) ([]db.Shelf, error)
	Update(ctx context.Context, id int64, name string) (*db.Shelf, error)
	Delete(ctx context.Context, id int64) error
	GetDefaultsByUser(ctx context.Context, userID int64) ([]db.Shelf, error)
}

// ShelfBookRepository defines operations for managing books on shelves and reading progress.
type ShelfBookRepository interface {
	Add(ctx context.Context, shelfID, bookID, pagesRead, progressPercent int64) (*db.ShelfBook, error)
	Get(ctx context.Context, shelfID, bookID int64) (*db.ShelfBook, error)
	UpdateProgress(ctx context.Context, id, pagesRead, progressPercent int64) (*db.ShelfBook, error)
	Remove(ctx context.Context, shelfID, bookID int64) error
	ListBooksOnShelf(ctx context.Context, shelfID int64) ([]db.ListBooksOnShelfRow, error)
	GetUserBookShelves(ctx context.Context, userID, bookID int64) ([]db.Shelf, error)
	GetUserReadingStats(ctx context.Context, userID int64) (*db.GetUserReadingStatsRow, error)
}

// ChallengeRepository defines operations for yearly reading challenges.
type ChallengeRepository interface {
	Upsert(ctx context.Context, userID, year, goal int64) (*db.ReadingChallenge, error)
	Get(ctx context.Context, userID, year int64) (*db.ReadingChallenge, error)
	CountFinishedInYear(ctx context.Context, userID, year int64) (int64, error)
}

// ReviewRepository defines operations for book reviews.
type ReviewRepository interface {
	Upsert(ctx context.Context, bookID, userID, rating int64, body string) (*db.Review, error)
	Get(ctx context.Context, bookID, userID int64) (*db.Review, error)
	Delete(ctx context.Context, bookID, userID int64) error
	ListByBook(ctx context.Context, bookID, limit, offset int64) ([]db.ListReviewsByBookRow, error)
	Stats(ctx context.Context, bookIDs []int64) ([]db.GetReviewStatsRow, error)
}

// RecommendationRepository defines operations for book recommendations.
type RecommendationRepository interface {
	ForUser(ctx context.Context, userID, limit int64) ([]db.GetRecommendationsRow, error)
}
