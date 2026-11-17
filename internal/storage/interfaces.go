package storage

import (
	"context"
	"time"

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

// TagCount is a tag and the number of books carrying it.
type TagCount struct {
	Tag   string `json:"tag"`
	Books int64  `json:"books"`
}

// SimilarBook is a book sharing tags with another, with the shared tags sorted.
type SimilarBook struct {
	Book       db.Book
	SharedTags []string
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
	// ListByTag returns one page of the books carrying tag, newest first.
	ListByTag(ctx context.Context, tag string, limit, offset int64) ([]db.Book, error)
	// SetTags replaces all of a book's tags; tags must already be validated and unique.
	SetTags(ctx context.Context, bookID int64, tags []string) error
	// TagsFor returns the sorted tags of each book; books without tags are absent from the map.
	TagsFor(ctx context.Context, bookIDs []int64) (map[int64][]string, error)
	// ListTagCounts returns one page of the tags in use starting with prefix (empty means all),
	// most books first then alphabetical, plus the total number of matching tags.
	ListTagCounts(ctx context.Context, prefix string, limit, offset int64) ([]TagCount, int64, error)
	// ListUserTagCounts returns one page of the tags carried by the books on userID's shelves
	// with the number of those distinct books carrying each (most first then alphabetical),
	// plus the total number of such tags.
	ListUserTagCounts(ctx context.Context, userID, limit, offset int64) ([]TagCount, int64, error)
	// ListSimilar returns one page of the books sharing at least one tag with bookID (never
	// bookID itself), most shared tags first then by id, plus the total number of such books.
	ListSimilar(ctx context.Context, bookID, limit, offset int64) ([]SimilarBook, int64, error)
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
	// Move re-homes a book's entry to another shelf in one transaction, keeping its progress.
	Move(ctx context.Context, fromShelfID, bookID, toShelfID int64) (*db.ListBooksOnShelfRow, error)
	ListBooksOnShelf(ctx context.Context, shelfID int64) ([]db.ListBooksOnShelfRow, error)
	GetUserBookShelves(ctx context.Context, userID, bookID int64) ([]db.Shelf, error)
	GetUserReadingStats(ctx context.Context, userID int64) (*db.GetUserReadingStatsRow, error)
	// ListMonthlyFinished returns, for each month of the year with finished books, the books
	// finished and their summed page counts. A book counts once, in the month of its earliest
	// finished_at that year. Months without finished books are omitted.
	ListMonthlyFinished(ctx context.Context, userID, year int64) ([]db.ListMonthlyFinishedBooksRow, error)
	// ListForExport returns every shelved book of the user, ordered by shelf name then book title.
	ListForExport(ctx context.Context, userID int64) ([]ShelfExportRow, error)
}

// FinishedBook is a book a reading challenge counts, with the earliest time it
// was finished in that year.
type FinishedBook struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Author     string    `json:"author"`
	FinishedAt time.Time `json:"finished_at"`
}

// ChallengeRepository defines operations for yearly reading challenges.
type ChallengeRepository interface {
	Upsert(ctx context.Context, userID, year, goal int64) (*db.ReadingChallenge, error)
	Get(ctx context.Context, userID, year int64) (*db.ReadingChallenge, error)
	ListByUser(ctx context.Context, userID int64) ([]db.ReadingChallenge, error)
	CountFinishedInYear(ctx context.Context, userID, year int64) (int64, error)
	ListFinishedInYear(ctx context.Context, userID, year int64) ([]FinishedBook, error)
	Delete(ctx context.Context, userID, year int64) error
}

// ReviewRepository defines operations for book reviews.
type ReviewRepository interface {
	Upsert(ctx context.Context, bookID, userID, rating int64, body string) (*db.Review, error)
	Get(ctx context.Context, bookID, userID int64) (*db.Review, error)
	Delete(ctx context.Context, bookID, userID int64) error
	ListByBook(
		ctx context.Context, bookID, minRating int64, sort string, limit, offset int64,
	) ([]db.ListReviewsByBookRow, error)
	ListByUser(ctx context.Context, userID, limit, offset int64) ([]db.ListReviewsByUserRow, error)
	Stats(ctx context.Context, bookIDs []int64) ([]db.GetReviewStatsRow, error)
	// RatingCounts returns a book's review count per rating, for ratings that have any.
	RatingCounts(ctx context.Context, bookID int64) ([]db.GetRatingCountsRow, error)
}

// RecommendationRepository defines operations for book recommendations.
type RecommendationRepository interface {
	ForUser(ctx context.Context, userID, limit int64) ([]db.GetRecommendationsRow, error)
}

// ShelfExportRow is one book on one of a user's shelves, as exported.
type ShelfExportRow struct {
	Shelf           string
	Title           string
	Author          string
	ISBN            string
	PagesRead       int64
	ProgressPercent int64
	FinishedAt      *time.Time
	AddedAt         time.Time
}
