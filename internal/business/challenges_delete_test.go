package business_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"goodreads-clone.com/internal/business"
	"goodreads-clone.com/internal/db"
	"goodreads-clone.com/internal/mocks"
	"goodreads-clone.com/internal/storage"
)

func TestBusiness_DeleteReadingChallenge(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	newBiz := func(t *testing.T) (*business.Business, *mocks.MockUserRepository, *mocks.MockChallengeRepository) {
		users := mocks.NewMockUserRepository(t)
		challenges := mocks.NewMockChallengeRepository(t)
		return &business.Business{Storage: &storage.Storage{Users: users, Challenges: challenges}}, users, challenges
	}

	b, users, challenges := newBiz(t)
	users.EXPECT().GetByID(ctx, int64(1)).Return(&db.User{ID: 1}, nil)
	challenges.EXPECT().Delete(ctx, int64(1), int64(2026)).Return(nil)
	require.NoError(t, b.DeleteReadingChallenge(ctx, 1, 2026))

	b, users, _ = newBiz(t)
	users.EXPECT().GetByID(ctx, int64(9)).Return(nil, sql.ErrNoRows)
	require.ErrorIs(t, b.DeleteReadingChallenge(ctx, 9, 2026), sql.ErrNoRows)

	b, users, challenges = newBiz(t)
	users.EXPECT().GetByID(ctx, int64(1)).Return(&db.User{ID: 1}, nil)
	challenges.EXPECT().Delete(ctx, int64(1), int64(2020)).Return(sql.ErrNoRows)
	require.ErrorIs(t, b.DeleteReadingChallenge(ctx, 1, 2020), sql.ErrNoRows)
}
