package storage

import (
	"context"
	"database/sql"
	"fmt"

	"goodreads-clone.com/internal/db"
)

type challengeRepository struct {
	queries *db.Queries
}

func (r *challengeRepository) Upsert(ctx context.Context, userID, year, goal int64) (*db.ReadingChallenge, error) {
	c, err := r.queries.UpsertReadingChallenge(ctx, db.UpsertReadingChallengeParams{
		UserID: userID,
		Year:   year,
		Goal:   goal,
	})
	if err != nil {
		return nil, fmt.Errorf("upsert reading challenge query failed: %w", err)
	}
	return &c, nil
}

func (r *challengeRepository) Get(ctx context.Context, userID, year int64) (*db.ReadingChallenge, error) {
	c, err := r.queries.GetReadingChallenge(ctx, db.GetReadingChallengeParams{UserID: userID, Year: year})
	if err != nil {
		return nil, fmt.Errorf("get reading challenge query failed: %w", err)
	}
	return &c, nil
}

func (r *challengeRepository) CountFinishedInYear(ctx context.Context, userID, year int64) (int64, error) {
	n, err := r.queries.CountBooksFinishedInYear(ctx, db.CountBooksFinishedInYearParams{UserID: userID, Year: year})
	if err != nil {
		return 0, fmt.Errorf("count books finished in year query failed: %w", err)
	}
	return n, nil
}

// Delete removes a challenge; it returns an error wrapping sql.ErrNoRows when none existed.
func (r *challengeRepository) Delete(ctx context.Context, userID, year int64) error {
	n, err := r.queries.DeleteReadingChallenge(ctx, db.DeleteReadingChallengeParams{UserID: userID, Year: year})
	if err != nil {
		return fmt.Errorf("delete reading challenge query failed: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("reading challenge not found: %w", sql.ErrNoRows)
	}
	return nil
}
