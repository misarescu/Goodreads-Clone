package storage

import (
	"context"
	"database/sql"
	"fmt"

	"goodreads-clone.com/internal/db"
)

type reviewRepository struct {
	queries *db.Queries
}

func (r *reviewRepository) Upsert(ctx context.Context, bookID, userID, rating int64, body string) (*db.Review, error) {
	rev, err := r.queries.UpsertReview(ctx, db.UpsertReviewParams{
		BookID: bookID, UserID: userID, Rating: rating, Body: body,
	})
	if err != nil {
		return nil, fmt.Errorf("upsert review query failed: %w", err)
	}
	return &rev, nil
}

func (r *reviewRepository) Get(ctx context.Context, bookID, userID int64) (*db.Review, error) {
	rev, err := r.queries.GetReview(ctx, db.GetReviewParams{BookID: bookID, UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("get review query failed: %w", err)
	}
	return &rev, nil
}

// Delete removes a review; it returns an error wrapping sql.ErrNoRows when none existed.
func (r *reviewRepository) Delete(ctx context.Context, bookID, userID int64) error {
	n, err := r.queries.DeleteReview(ctx, db.DeleteReviewParams{BookID: bookID, UserID: userID})
	if err != nil {
		return fmt.Errorf("delete review query failed: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("review not found: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *reviewRepository) ListByBook(
	ctx context.Context, bookID, minRating int64, sort string, limit, offset int64,
) ([]db.ListReviewsByBookRow, error) {
	var rows []db.ListReviewsByBookRow
	if sort == "rating" {
		byRating, err := r.queries.ListReviewsByBookRating(ctx, db.ListReviewsByBookRatingParams{
			BookID: bookID, MinRating: minRating, Limit: limit, Offset: offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list reviews by rating query failed: %w", err)
		}
		rows = make([]db.ListReviewsByBookRow, len(byRating))
		for i, row := range byRating {
			rows[i] = db.ListReviewsByBookRow(row)
		}
		return rows, nil
	}
	rows, err := r.queries.ListReviewsByBook(ctx, db.ListReviewsByBookParams{
		BookID: bookID, MinRating: minRating, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list reviews query failed: %w", err)
	}
	return rows, nil
}

func (r *reviewRepository) Stats(ctx context.Context, bookIDs []int64) ([]db.GetReviewStatsRow, error) {
	if len(bookIDs) == 0 {
		return nil, nil
	}
	rows, err := r.queries.GetReviewStats(ctx, bookIDs)
	if err != nil {
		return nil, fmt.Errorf("review stats query failed: %w", err)
	}
	return rows, nil
}

func (r *reviewRepository) ListByUser(
	ctx context.Context, userID, limit, offset int64,
) ([]db.ListReviewsByUserRow, error) {
	rows, err := r.queries.ListReviewsByUser(ctx, db.ListReviewsByUserParams{UserID: userID, Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list user reviews query failed: %w", err)
	}
	return rows, nil
}

func (r *reviewRepository) RatingCounts(ctx context.Context, bookID int64) ([]db.GetRatingCountsRow, error) {
	rows, err := r.queries.GetRatingCounts(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("rating counts query failed: %w", err)
	}
	return rows, nil
}
