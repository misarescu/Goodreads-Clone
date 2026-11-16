package business

import (
	"context"
	"errors"
	"fmt"
	"time"

	"goodreads-clone.com/internal/db"
	"goodreads-clone.com/internal/storage"
)

// Reading challenge goal bounds.
const (
	MinChallengeGoal = 1
	MaxChallengeGoal = 1000
)

// ErrInvalidGoal is returned when a challenge goal is outside the allowed range.
var ErrInvalidGoal = errors.New("goal must be between 1 and 1000")

// ChallengeProgress is a reading challenge with the user's progress so far.
type ChallengeProgress struct {
	Year          int64 `json:"year"`
	Goal          int64 `json:"goal"`
	Finished      int64 `json:"finished"`
	ExpectedByNow int64 `json:"expected_by_now"`
	AheadBy       int64 `json:"ahead_by"`
}

// ChallengeDetail is a challenge's progress with the books it counts.
type ChallengeDetail struct {
	ChallengeProgress

	Books []storage.FinishedBook `json:"books"`
}

func (b *Business) now() time.Time {
	if b.Now != nil {
		return b.Now()
	}
	return time.Now()
}

// SetReadingChallenge creates or updates a user's challenge for a year. It
// returns an error wrapping sql.ErrNoRows when the user does not exist.
func (b *Business) SetReadingChallenge(ctx context.Context, userID, year, goal int64) (*db.ReadingChallenge, error) {
	if goal < MinChallengeGoal || goal > MaxChallengeGoal {
		return nil, ErrInvalidGoal
	}
	if _, err := b.Storage.Users.GetByID(ctx, userID); err != nil {
		return nil, fmt.Errorf("failed to get user for reading challenge: %w", err)
	}
	c, err := b.Storage.Challenges.Upsert(ctx, userID, year, goal)
	if err != nil {
		return nil, fmt.Errorf("failed to set reading challenge: %w", err)
	}
	return c, nil
}

// GetReadingChallengeProgress returns a challenge with progress and the books it counts. It returns an
// error wrapping sql.ErrNoRows when the user or the challenge does not exist.
func (b *Business) GetReadingChallengeProgress(ctx context.Context, userID, year int64) (*ChallengeDetail, error) {
	if _, err := b.Storage.Users.GetByID(ctx, userID); err != nil {
		return nil, fmt.Errorf("failed to get user for reading challenge: %w", err)
	}
	c, err := b.Storage.Challenges.Get(ctx, userID, year)
	if err != nil {
		return nil, fmt.Errorf("failed to get reading challenge: %w", err)
	}
	books, err := b.Storage.Challenges.ListFinishedInYear(ctx, userID, year)
	if err != nil {
		return nil, fmt.Errorf("failed to list finished books: %w", err)
	}
	// Finished comes from the same list, so len(Books) == Finished.
	return &ChallengeDetail{
		ChallengeProgress: b.progressFor(c, int64(len(books))),
		Books:             books,
	}, nil
}

// ListReadingChallengeProgress returns all of a user's challenges with
// progress, newest year first. It returns an error wrapping sql.ErrNoRows when
// the user does not exist.
func (b *Business) ListReadingChallengeProgress(ctx context.Context, userID int64) ([]ChallengeProgress, error) {
	if _, err := b.Storage.Users.GetByID(ctx, userID); err != nil {
		return nil, fmt.Errorf("failed to get user for reading challenges: %w", err)
	}
	challenges, err := b.Storage.Challenges.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list reading challenges: %w", err)
	}
	out := make([]ChallengeProgress, 0, len(challenges))
	for i := range challenges {
		p, err := b.challengeProgress(ctx, &challenges[i])
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, nil
}

func (b *Business) challengeProgress(ctx context.Context, c *db.ReadingChallenge) (*ChallengeProgress, error) {
	finished, err := b.Storage.Challenges.CountFinishedInYear(ctx, c.UserID, c.Year)
	if err != nil {
		return nil, fmt.Errorf("failed to count finished books: %w", err)
	}
	p := b.progressFor(c, finished)
	return &p, nil
}

func (b *Business) progressFor(c *db.ReadingChallenge, finished int64) ChallengeProgress {
	expected := expectedByNow(c.Goal, c.Year, b.now())
	return ChallengeProgress{
		Year:          c.Year,
		Goal:          c.Goal,
		Finished:      finished,
		ExpectedByNow: expected,
		AheadBy:       finished - expected,
	}
}

// DeleteReadingChallenge removes a user's challenge for a year. It returns an
// error wrapping sql.ErrNoRows when the user or the challenge does not exist.
func (b *Business) DeleteReadingChallenge(ctx context.Context, userID, year int64) error {
	if _, err := b.Storage.Users.GetByID(ctx, userID); err != nil {
		return fmt.Errorf("failed to get user for reading challenge: %w", err)
	}
	if err := b.Storage.Challenges.Delete(ctx, userID, year); err != nil {
		return fmt.Errorf("failed to delete reading challenge: %w", err)
	}
	return nil
}

// expectedByNow is goal times the fraction of the (UTC) year elapsed at now,
// rounded down: the whole goal for a past year, 0 for a future one.
func expectedByNow(goal, year int64, now time.Time) int64 {
	now = now.UTC()
	start := time.Date(int(year), time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(int(year)+1, time.January, 1, 0, 0, 0, 0, time.UTC)
	switch {
	case !now.After(start):
		return 0
	case !now.Before(end):
		return goal
	}
	return goal * int64(now.Sub(start)/time.Second) / int64(end.Sub(start)/time.Second)
}
