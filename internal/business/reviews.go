package business

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"goodreads-clone.com/internal/db"
)

// Review bounds.
const (
	MinRating     = 1
	MaxRating     = 5
	MaxReviewBody = 5000
)

// Review list orders.
const (
	SortRecent = "recent"
	SortRating = "rating"
)

// Review validation errors.
var (
	ErrInvalidRating    = errors.New("rating must be between 1 and 5")
	ErrInvalidBody      = errors.New("body must be between 1 and 5000 characters")
	ErrInvalidSort      = errors.New("sort must be recent or rating")
	ErrInvalidMinRating = errors.New("min_rating must be an integer from 1 to 5")
)

// BookWithStats is a book with its review statistics.
type BookWithStats struct {
	db.Book

	AverageRating *float64 `json:"average_rating"`
	ReviewCount   int64    `json:"review_count"`
}

// SaveReview creates the user's review of a book or replaces the existing one.
// created reports whether a new review was made. It returns an error wrapping
// sql.ErrNoRows when the book or the user does not exist.
func (b *Business) SaveReview(
	ctx context.Context, bookID, userID, rating int64, body string,
) (rev *db.Review, created bool, err error) {
	if rating < MinRating || rating > MaxRating {
		return nil, false, ErrInvalidRating
	}
	if strings.TrimSpace(body) == "" || utf8.RuneCountInString(body) > MaxReviewBody {
		return nil, false, ErrInvalidBody
	}
	if _, err := b.Storage.Books.GetByID(ctx, bookID); err != nil {
		return nil, false, fmt.Errorf("failed to get book for review: %w", err)
	}
	if _, err := b.Storage.Users.GetByID(ctx, userID); err != nil {
		return nil, false, fmt.Errorf("failed to get user for review: %w", err)
	}
	_, err = b.Storage.Reviews.Get(ctx, bookID, userID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		created = true
	case err != nil:
		return nil, false, fmt.Errorf("failed to look up review: %w", err)
	}
	rev, err = b.Storage.Reviews.Upsert(ctx, bookID, userID, rating, body)
	if err != nil {
		return nil, false, fmt.Errorf("failed to save review: %w", err)
	}
	return rev, created, nil
}

// ListReviews returns a book's reviews ordered by sort (SortRecent or SortRating)
// and rated at least minRating (0 means any). It returns ErrInvalidSort or
// ErrInvalidMinRating for bad options, and an error wrapping sql.ErrNoRows
// when the book does not exist.
func (b *Business) ListReviews(
	ctx context.Context, bookID, minRating int64, sort string, limit, offset int64,
) ([]db.ListReviewsByBookRow, error) {
	if sort != SortRecent && sort != SortRating {
		return nil, ErrInvalidSort
	}
	if minRating != 0 && (minRating < MinRating || minRating > MaxRating) {
		return nil, ErrInvalidMinRating
	}
	if _, err := b.Storage.Books.GetByID(ctx, bookID); err != nil {
		return nil, fmt.Errorf("failed to get book for reviews: %w", err)
	}
	rows, err := b.Storage.Reviews.ListByBook(ctx, bookID, minRating, sort, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list reviews: %w", err)
	}
	return rows, nil
}

// DeleteReview removes a user's review of a book. It returns an error
// wrapping sql.ErrNoRows when there is no such review.
func (b *Business) DeleteReview(ctx context.Context, bookID, userID int64) error {
	if err := b.Storage.Reviews.Delete(ctx, bookID, userID); err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}
	return nil
}

// WithReviewStats attaches the average rating (one decimal, nil without
// reviews) and review count to each book.
func (b *Business) WithReviewStats(ctx context.Context, books []db.Book) ([]BookWithStats, error) {
	ids := make([]int64, len(books))
	for i, bk := range books {
		ids[i] = bk.ID
	}
	stats, err := b.Storage.Reviews.Stats(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get review stats: %w", err)
	}
	byBook := make(map[int64]db.GetReviewStatsRow, len(stats))
	for _, s := range stats {
		byBook[s.BookID] = s
	}
	out := make([]BookWithStats, len(books))
	for i, bk := range books {
		out[i] = BookWithStats{Book: bk}
		if s, ok := byBook[bk.ID]; ok && s.ReviewCount > 0 {
			avg := math.Round(s.AverageRating*10) / 10
			out[i].AverageRating = &avg
			out[i].ReviewCount = s.ReviewCount
		}
	}
	return out, nil
}

// WithReviewStat is WithReviewStats for a single book.
func (b *Business) WithReviewStat(ctx context.Context, book *db.Book) (*BookWithStats, error) {
	out, err := b.WithReviewStats(ctx, []db.Book{*book})
	if err != nil {
		return nil, err
	}
	return &out[0], nil
}

// ListUserReviews returns the reviews a user wrote, newest first, each with
// the reviewed book's title. It returns an error wrapping sql.ErrNoRows when
// the user does not exist.
func (b *Business) ListUserReviews(
	ctx context.Context, userID, limit, offset int64,
) ([]db.ListReviewsByUserRow, error) {
	if _, err := b.Storage.Users.GetByID(ctx, userID); err != nil {
		return nil, fmt.Errorf("failed to get user for reviews: %w", err)
	}
	rows, err := b.Storage.Reviews.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user reviews: %w", err)
	}
	return rows, nil
}
