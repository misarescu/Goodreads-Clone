package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChallengeRepository_ListByUser(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	user := createTestUser(t, s, "a@example.com", "alice")
	other := createTestUser(t, s, "b@example.com", "bob")
	for _, c := range []struct{ user, year int64 }{{user.ID, 2025}, {user.ID, 2027}, {user.ID, 2026}, {other.ID, 2028}} {
		_, err := s.Challenges.Upsert(ctx, c.user, c.year, 10)
		require.NoError(t, err)
	}

	rows, err := s.Challenges.ListByUser(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, rows, 3)
	assert.Equal(t, []int64{2027, 2026, 2025}, []int64{rows[0].Year, rows[1].Year, rows[2].Year})

	rows, err = s.Challenges.ListByUser(ctx, 999)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestChallengeRepository_Delete(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	user := createTestUser(t, s, "a@example.com", "alice")
	other := createTestUser(t, s, "b@example.com", "bob")
	for _, c := range []struct{ user, year int64 }{{user.ID, 2026}, {user.ID, 2027}, {other.ID, 2026}} {
		_, err := s.Challenges.Upsert(ctx, c.user, c.year, 10)
		require.NoError(t, err)
	}

	require.NoError(t, s.Challenges.Delete(ctx, user.ID, 2026))

	_, err := s.Challenges.Get(ctx, user.ID, 2026)
	require.ErrorIs(t, err, sql.ErrNoRows)
	_, err = s.Challenges.Get(ctx, user.ID, 2027)
	require.NoError(t, err)
	_, err = s.Challenges.Get(ctx, other.ID, 2026)
	require.NoError(t, err)

	require.ErrorIs(t, s.Challenges.Delete(ctx, user.ID, 2026), sql.ErrNoRows)
}

func TestChallengeRepository_ListFinishedInYear(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	user := createTestUser(t, s, "a@example.com", "alice")
	other := createTestUser(t, s, "b@example.com", "bob")
	shelf1 := createTestShelf(t, s, user.ID, "One", false)
	shelf2 := createTestShelf(t, s, user.ID, "Two", false)
	otherShelf := createTestShelf(t, s, other.ID, "Theirs", false)
	multi := createTestBook(t, s, "Multi", "A")
	late := createTestBook(t, s, "Late", "B")
	lastYear := createTestBook(t, s, "Last year", "C")
	unfinished := createTestBook(t, s, "Unfinished", "D")
	theirs := createTestBook(t, s, "Theirs", "E")

	for _, r := range []struct {
		shelf, book int64
		at          string
	}{
		{shelf1.ID, multi.ID, "2026-03-05 10:00:00"},
		{shelf2.ID, multi.ID, "2026-02-01 10:00:00"},
		{shelf1.ID, late.ID, "2026-11-01 08:30:00"},
		{shelf1.ID, lastYear.ID, "2025-12-31 23:59:59"},
		{shelf1.ID, unfinished.ID, ""},
		{otherShelf.ID, theirs.ID, "2026-04-01 10:00:00"},
	} {
		_, err := s.ShelfBooks.Add(ctx, r.shelf, r.book, 0, 0)
		require.NoError(t, err)
		if r.at != "" {
			_, err = s.DB().ExecContext(ctx, `UPDATE shelf_books SET finished_at = ? WHERE shelf_id = ? AND book_id = ?`,
				r.at, r.shelf, r.book)
			require.NoError(t, err)
		}
	}

	got, err := s.Challenges.ListFinishedInYear(ctx, user.ID, 2026)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, late.ID, got[0].ID)
	assert.Equal(t, "Late", got[0].Title)
	assert.Equal(t, "B", got[0].Author)
	assert.Equal(t, time.Date(2026, 11, 1, 8, 30, 0, 0, time.UTC), got[0].FinishedAt)
	assert.Equal(t, multi.ID, got[1].ID)
	assert.Equal(t, time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC), got[1].FinishedAt)

	n, err := s.Challenges.CountFinishedInYear(ctx, user.ID, 2026)
	require.NoError(t, err)
	assert.Equal(t, int64(len(got)), n)

	got, err = s.Challenges.ListFinishedInYear(ctx, user.ID, 2024)
	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Empty(t, got)
}
