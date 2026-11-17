package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReviewRepository_ListByUser(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	user := createTestUser(t, s, "a@example.com", "alice")
	other := createTestUser(t, s, "b@example.com", "bob")
	b1 := createTestBook(t, s, "First", "A")
	b2 := createTestBook(t, s, "Second", "A")

	for _, q := range []struct {
		book, user int64
		at         string
	}{
		{b1.ID, user.ID, "2026-01-01 10:00:00"},
		{b2.ID, user.ID, "2026-02-01 10:00:00"},
		{b1.ID, other.ID, "2026-03-01 10:00:00"},
	} {
		_, err := s.DB().ExecContext(ctx,
			`INSERT INTO reviews (book_id, user_id, rating, body, created_at) VALUES (?, ?, 4, 'x', ?)`, q.book, q.user, q.at)
		require.NoError(t, err)
	}

	rows, err := s.Reviews.ListByUser(ctx, user.ID, 20, 0)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "Second", rows[0].BookTitle)
	assert.Equal(t, "First", rows[1].BookTitle)

	rows, err = s.Reviews.ListByUser(ctx, user.ID, 1, 1)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "First", rows[0].BookTitle)

	rows, err = s.Reviews.ListByUser(ctx, 999, 20, 0)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestReviewRepository_RatingCounts(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	book := createTestBook(t, s, "First", "A")
	for i, rating := range []int{4, 4, 5} {
		u := createTestUser(t, s, fmt.Sprintf("u%d@example.com", i), fmt.Sprintf("u%d", i))
		_, err := s.Reviews.Upsert(ctx, book.ID, u.ID, int64(rating), "x")
		require.NoError(t, err)
	}

	rows, err := s.Reviews.RatingCounts(ctx, book.ID)
	require.NoError(t, err)
	got := map[int64]int64{}
	for _, r := range rows {
		got[r.Rating] = r.ReviewCount
	}
	assert.Equal(t, map[int64]int64{4: 2, 5: 1}, got)

	rows, err = s.Reviews.RatingCounts(ctx, 999)
	require.NoError(t, err)
	assert.Empty(t, rows)
}
