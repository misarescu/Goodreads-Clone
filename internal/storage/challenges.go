package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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

func (r *challengeRepository) ListByUser(ctx context.Context, userID int64) ([]db.ReadingChallenge, error) {
	rows, err := r.queries.ListReadingChallengesByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list reading challenges query failed: %w", err)
	}
	return rows, nil
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

// finishedAtLayouts are the formats a finished_at timestamp can be stored in.
var finishedAtLayouts = []string{"2006-01-02 15:04:05", time.RFC3339Nano, "2006-01-02T15:04:05", "2006-01-02"}

func parseFinishedAt(v string) (time.Time, error) {
	for _, layout := range finishedAtLayouts {
		if t, err := time.Parse(layout, v); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised finished_at %q", v)
}

// ListFinishedInYear returns the distinct books on the user's shelves finished
// in the year (the books CountFinishedInYear counts), newest first, each with
// its earliest finished_at that year.
func (r *challengeRepository) ListFinishedInYear(ctx context.Context, userID, year int64) ([]FinishedBook, error) {
	rows, err := r.queries.ListBooksFinishedInYear(ctx, db.ListBooksFinishedInYearParams{UserID: userID, Year: year})
	if err != nil {
		return nil, fmt.Errorf("list books finished in year query failed: %w", err)
	}
	out := make([]FinishedBook, 0, len(rows))
	for _, row := range rows {
		at, err := parseFinishedAt(row.FinishedAt)
		if err != nil {
			return nil, fmt.Errorf("list books finished in year: %w", err)
		}
		out = append(out, FinishedBook{ID: row.ID, Title: row.Title, Author: row.Author, FinishedAt: at})
	}
	return out, nil
}
