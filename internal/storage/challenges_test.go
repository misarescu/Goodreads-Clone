package storage

import (
	"context"
	"database/sql"
	"testing"

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
