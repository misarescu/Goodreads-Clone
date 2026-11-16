package business_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goodreads-clone.com/internal/business"
	"goodreads-clone.com/internal/db"
	"goodreads-clone.com/internal/mocks"
	"goodreads-clone.com/internal/storage"
)

func newChallengeBusiness(
	t *testing.T,
) (*business.Business, *mocks.MockUserRepository, *mocks.MockChallengeRepository) {
	users := mocks.NewMockUserRepository(t)
	ch := mocks.NewMockChallengeRepository(t)
	b := &business.Business{
		Storage: &storage.Storage{Users: users, Challenges: ch},
		Now:     func() time.Time { return time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC) },
	}
	return b, users, ch
}

func TestBusiness_ListReadingChallengeProgress(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b, users, ch := newChallengeBusiness(t)
	users.EXPECT().GetByID(ctx, int64(1)).Return(&db.User{ID: 1}, nil)
	ch.EXPECT().ListByUser(ctx, int64(1)).Return([]db.ReadingChallenge{
		{UserID: 1, Year: 2026, Goal: 25},
		{UserID: 1, Year: 2025, Goal: 10},
	}, nil)
	ch.EXPECT().CountFinishedInYear(ctx, int64(1), int64(2026)).Return(int64(2), nil)
	ch.EXPECT().CountFinishedInYear(ctx, int64(1), int64(2025)).Return(int64(4), nil)

	got, err := b.ListReadingChallengeProgress(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, []business.ChallengeProgress{
		{Year: 2026, Goal: 25, Finished: 2, ExpectedByNow: 12, AheadBy: -10},
		{Year: 2025, Goal: 10, Finished: 4, ExpectedByNow: 10, AheadBy: -6},
	}, got)
}

func TestBusiness_ListReadingChallengeProgress_Empty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b, users, ch := newChallengeBusiness(t)
	users.EXPECT().GetByID(ctx, int64(1)).Return(&db.User{ID: 1}, nil)
	ch.EXPECT().ListByUser(ctx, int64(1)).Return(nil, nil)

	got, err := b.ListReadingChallengeProgress(ctx, 1)
	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Empty(t, got)
}

func TestBusiness_ListReadingChallengeProgress_Errors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("unknown user", func(t *testing.T) {
		b, users, _ := newChallengeBusiness(t)
		users.EXPECT().GetByID(ctx, int64(9)).Return(nil, sql.ErrNoRows)
		_, err := b.ListReadingChallengeProgress(ctx, 9)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
	t.Run("list fails", func(t *testing.T) {
		b, users, ch := newChallengeBusiness(t)
		boom := errors.New("boom")
		users.EXPECT().GetByID(ctx, int64(1)).Return(&db.User{ID: 1}, nil)
		ch.EXPECT().ListByUser(ctx, int64(1)).Return(nil, boom)
		_, err := b.ListReadingChallengeProgress(ctx, 1)
		assert.ErrorIs(t, err, boom)
	})
	t.Run("count fails", func(t *testing.T) {
		b, users, ch := newChallengeBusiness(t)
		boom := errors.New("boom")
		users.EXPECT().GetByID(ctx, int64(1)).Return(&db.User{ID: 1}, nil)
		ch.EXPECT().ListByUser(ctx, int64(1)).Return([]db.ReadingChallenge{{UserID: 1, Year: 2026, Goal: 5}}, nil)
		ch.EXPECT().CountFinishedInYear(ctx, int64(1), int64(2026)).Return(int64(0), boom)
		_, err := b.ListReadingChallengeProgress(ctx, 1)
		assert.ErrorIs(t, err, boom)
	})
}

func TestBusiness_GetReadingChallengeProgress_Books(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b, users, ch := newChallengeBusiness(t)
	users.EXPECT().GetByID(ctx, int64(1)).Return(&db.User{ID: 1}, nil)
	ch.EXPECT().Get(ctx, int64(1), int64(2026)).Return(&db.ReadingChallenge{UserID: 1, Year: 2026, Goal: 25}, nil)
	books := []storage.FinishedBook{
		{ID: 2, Title: "B", Author: "Y", FinishedAt: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 1, Title: "A", Author: "X", FinishedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)},
	}
	ch.EXPECT().ListFinishedInYear(ctx, int64(1), int64(2026)).Return(books, nil)

	got, err := b.GetReadingChallengeProgress(ctx, 1, 2026)
	require.NoError(t, err)
	assert.Equal(t, books, got.Books)
	assert.Equal(t, int64(len(got.Books)), got.Finished)
	assert.Equal(t, int64(-10), got.AheadBy)
}

func TestBusiness_GetReadingChallengeProgress_BooksError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b, users, ch := newChallengeBusiness(t)
	users.EXPECT().GetByID(ctx, int64(1)).Return(&db.User{ID: 1}, nil)
	ch.EXPECT().Get(ctx, int64(1), int64(2026)).Return(&db.ReadingChallenge{UserID: 1, Year: 2026, Goal: 25}, nil)
	ch.EXPECT().ListFinishedInYear(ctx, int64(1), int64(2026)).Return(nil, errors.New("boom"))

	_, err := b.GetReadingChallengeProgress(ctx, 1, 2026)
	require.Error(t, err)
}
