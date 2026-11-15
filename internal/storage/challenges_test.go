package storage

import (
	"context"
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
