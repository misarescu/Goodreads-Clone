package business_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goodreads-clone.com/internal/business"
	"goodreads-clone.com/internal/db"
	"goodreads-clone.com/internal/mocks"
	"goodreads-clone.com/internal/storage"
)

type reviewMocks struct {
	books   *mocks.MockBookRepository
	users   *mocks.MockUserRepository
	reviews *mocks.MockReviewRepository
}

func newReviewBusiness(t *testing.T) (*business.Business, reviewMocks) {
	books := mocks.NewMockBookRepository(t)
	users := mocks.NewMockUserRepository(t)
	reviews := mocks.NewMockReviewRepository(t)
	store := &storage.Storage{Books: books, Users: users, Reviews: reviews}
	return &business.Business{Storage: store}, reviewMocks{books, users, reviews}
}

func TestBusiness_SaveReview_Validation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		rating int64
		body   string
		want   error
	}{
		{"rating 0", 0, "x", business.ErrInvalidRating},
		{"rating 6", 6, "x", business.ErrInvalidRating},
		{"empty body", 3, "", business.ErrInvalidBody},
		{"blank body", 3, "  \n", business.ErrInvalidBody},
		{"long body", 3, strings.Repeat("a", 5001), business.ErrInvalidBody},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := newReviewBusiness(t)
			_, _, err := b.SaveReview(context.Background(), 1, 1, tt.rating, tt.body)
			assert.ErrorIs(t, err, tt.want)
		})
	}
}

func TestBusiness_SaveReview(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	saved := &db.Review{ID: 1, BookID: 1, UserID: 2, Rating: 4, Body: "ok"}

	t.Run("creates", func(t *testing.T) {
		b, m := newReviewBusiness(t)
		books, users, reviews := m.books, m.users, m.reviews
		books.EXPECT().GetByID(ctx, int64(1)).Return(&db.Book{ID: 1}, nil)
		users.EXPECT().GetByID(ctx, int64(2)).Return(&db.User{ID: 2}, nil)
		reviews.EXPECT().Get(ctx, int64(1), int64(2)).Return(nil, sql.ErrNoRows)
		reviews.EXPECT().Upsert(ctx, int64(1), int64(2), int64(4), "ok").Return(saved, nil)
		rev, created, err := b.SaveReview(ctx, 1, 2, 4, "ok")
		require.NoError(t, err)
		assert.True(t, created)
		assert.Equal(t, saved, rev)
	})

	t.Run("replaces", func(t *testing.T) {
		b, m := newReviewBusiness(t)
		books, users, reviews := m.books, m.users, m.reviews
		books.EXPECT().GetByID(ctx, int64(1)).Return(&db.Book{ID: 1}, nil)
		users.EXPECT().GetByID(ctx, int64(2)).Return(&db.User{ID: 2}, nil)
		reviews.EXPECT().Get(ctx, int64(1), int64(2)).Return(saved, nil)
		reviews.EXPECT().Upsert(ctx, int64(1), int64(2), int64(4), "ok").Return(saved, nil)
		_, created, err := b.SaveReview(ctx, 1, 2, 4, "ok")
		require.NoError(t, err)
		assert.False(t, created)
	})

	t.Run("unknown book", func(t *testing.T) {
		b, m := newReviewBusiness(t)
		books := m.books
		books.EXPECT().GetByID(ctx, int64(1)).Return(nil, sql.ErrNoRows)
		_, _, err := b.SaveReview(ctx, 1, 2, 4, "ok")
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("unknown user", func(t *testing.T) {
		b, m := newReviewBusiness(t)
		books, users := m.books, m.users
		books.EXPECT().GetByID(ctx, int64(1)).Return(&db.Book{ID: 1}, nil)
		users.EXPECT().GetByID(ctx, int64(2)).Return(nil, sql.ErrNoRows)
		_, _, err := b.SaveReview(ctx, 1, 2, 4, "ok")
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("lookup failure", func(t *testing.T) {
		b, m := newReviewBusiness(t)
		books, users, reviews := m.books, m.users, m.reviews
		boom := errors.New("boom")
		books.EXPECT().GetByID(ctx, int64(1)).Return(&db.Book{ID: 1}, nil)
		users.EXPECT().GetByID(ctx, int64(2)).Return(&db.User{ID: 2}, nil)
		reviews.EXPECT().Get(ctx, int64(1), int64(2)).Return(nil, boom)
		_, _, err := b.SaveReview(ctx, 1, 2, 4, "ok")
		assert.ErrorIs(t, err, boom)
	})
}

func TestBusiness_ListAndDeleteReviews(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	b, m := newReviewBusiness(t)
	books, reviews := m.books, m.reviews
	books.EXPECT().GetByID(ctx, int64(1)).Return(&db.Book{ID: 1}, nil)
	rows0 := []db.ListReviewsByBookRow{{ID: 1, Username: "u"}}
	reviews.EXPECT().ListByBook(ctx, int64(1), int64(3), "rating", int64(20), int64(0)).Return(rows0, nil)
	rows, err := b.ListReviews(ctx, 1, 3, business.SortRating, 20, 0)
	require.NoError(t, err)
	assert.Len(t, rows, 1)

	b, m = newReviewBusiness(t)
	books = m.books
	books.EXPECT().GetByID(ctx, int64(9)).Return(nil, sql.ErrNoRows)
	_, err = b.ListReviews(ctx, 9, 0, business.SortRecent, 20, 0)
	require.ErrorIs(t, err, sql.ErrNoRows)

	b, _ = newReviewBusiness(t)
	_, err = b.ListReviews(ctx, 1, 0, "best", 20, 0)
	require.ErrorIs(t, err, business.ErrInvalidSort)
	for _, bad := range []int64{-1, 6} {
		_, err = b.ListReviews(ctx, 1, bad, business.SortRecent, 20, 0)
		require.ErrorIs(t, err, business.ErrInvalidMinRating)
	}

	b, m = newReviewBusiness(t)
	reviews = m.reviews
	reviews.EXPECT().Delete(ctx, int64(1), int64(2)).Return(sql.ErrNoRows)
	assert.ErrorIs(t, b.DeleteReview(ctx, 1, 2), sql.ErrNoRows)
}

func TestBusiness_WithReviewStats(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b, m := newReviewBusiness(t)
	reviews := m.reviews
	reviews.EXPECT().Stats(ctx, []int64{1, 2}).Return([]db.GetReviewStatsRow{
		{BookID: 2, ReviewCount: 3, AverageRating: 13.0 / 3},
	}, nil)

	got, err := b.WithReviewStats(ctx, []db.Book{{ID: 1}, {ID: 2}})
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Nil(t, got[0].AverageRating)
	assert.Zero(t, got[0].ReviewCount)
	require.NotNil(t, got[1].AverageRating)
	assert.InDelta(t, 4.3, *got[1].AverageRating, 1e-9)
	assert.Equal(t, int64(3), got[1].ReviewCount)
}

func TestBusiness_ListUserReviews(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	b, m := newReviewBusiness(t)
	m.users.EXPECT().GetByID(ctx, int64(1)).Return(&db.User{ID: 1}, nil)
	m.reviews.EXPECT().ListByUser(ctx, int64(1), int64(20), int64(5)).
		Return([]db.ListReviewsByUserRow{{ID: 1, BookTitle: "T"}}, nil)
	rows, err := b.ListUserReviews(ctx, 1, 20, 5)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "T", rows[0].BookTitle)

	b, m = newReviewBusiness(t)
	m.users.EXPECT().GetByID(ctx, int64(9)).Return(nil, sql.ErrNoRows)
	_, err = b.ListUserReviews(ctx, 9, 20, 0)
	require.ErrorIs(t, err, sql.ErrNoRows)

	b, m = newReviewBusiness(t)
	boom := errors.New("boom")
	m.users.EXPECT().GetByID(ctx, int64(1)).Return(&db.User{ID: 1}, nil)
	m.reviews.EXPECT().ListByUser(ctx, int64(1), int64(20), int64(0)).Return(nil, boom)
	_, err = b.ListUserReviews(ctx, 1, 20, 0)
	require.ErrorIs(t, err, boom)
}
