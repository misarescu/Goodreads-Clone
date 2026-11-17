package business_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"goodreads-clone.com/internal/business"
	"goodreads-clone.com/internal/db"
	"goodreads-clone.com/internal/mocks"
	"goodreads-clone.com/internal/storage"
)

func newTagBusiness(t *testing.T) (*business.Business, *mocks.MockBookRepository) {
	books := mocks.NewMockBookRepository(t)
	return &business.Business{Storage: &storage.Storage{Books: books}}, books
}

func TestNormalizeTag(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{
		"sci-fi": "sci-fi", "Fantasy": "fantasy", "A1-b2": "a1-b2", "x": "x",
		strings.Repeat("a", 32): strings.Repeat("a", 32),
	} {
		got, err := business.NormalizeTag(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got)
	}
	for _, in := range []string{"", "   ", "sci fi", "sci_fi", "café", "a/b", "tag!", strings.Repeat("a", 33)} {
		_, err := business.NormalizeTag(in)
		require.ErrorIs(t, err, business.ErrInvalidTag, in)
		assert.Contains(t, err.Error(), `"`+in+`"`, "error names the tag")
	}
}

func TestBusiness_SetBookTags(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("normalizes, collapses duplicates and sorts", func(t *testing.T) {
		t.Parallel()
		b, books := newTagBusiness(t)
		books.EXPECT().GetByID(ctx, int64(1)).Return(&db.Book{ID: 1}, nil)
		books.EXPECT().SetTags(ctx, int64(1), []string{"fantasy", "sci-fi"}).Return(nil)
		got, err := b.SetBookTags(ctx, 1, []string{" Sci-Fi", "fantasy", "SCI-FI"})
		require.NoError(t, err)
		assert.Equal(t, []string{"fantasy", "sci-fi"}, got.Tags)
	})

	t.Run("empty list clears", func(t *testing.T) {
		t.Parallel()
		b, books := newTagBusiness(t)
		books.EXPECT().GetByID(ctx, int64(1)).Return(&db.Book{ID: 1}, nil)
		books.EXPECT().SetTags(ctx, int64(1), []string{}).Return(nil)
		got, err := b.SetBookTags(ctx, 1, nil)
		require.NoError(t, err)
		out, _ := json.Marshal(got)
		assert.Contains(t, string(out), `"tags":[]`)
	})

	t.Run("invalid tag stores nothing", func(t *testing.T) {
		t.Parallel()
		b, _ := newTagBusiness(t)
		_, err := b.SetBookTags(ctx, 1, []string{"ok", "not ok"})
		require.ErrorIs(t, err, business.ErrInvalidTag)
		assert.Contains(t, err.Error(), "not ok")
	})

	t.Run("limit is 10 distinct tags", func(t *testing.T) {
		t.Parallel()
		b, books := newTagBusiness(t)
		ten := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
		books.EXPECT().GetByID(ctx, int64(1)).Return(&db.Book{ID: 1}, nil)
		books.EXPECT().SetTags(ctx, int64(1), ten).Return(nil)
		_, err := b.SetBookTags(ctx, 1, append(append([]string{}, ten...), "A", "j"))
		require.NoError(t, err, "duplicates do not count toward the limit")

		_, err = b.SetBookTags(ctx, 1, append(append([]string{}, ten...), "k"))
		require.ErrorIs(t, err, business.ErrInvalidTag)
	})

	t.Run("unknown book", func(t *testing.T) {
		t.Parallel()
		b, books := newTagBusiness(t)
		books.EXPECT().GetByID(ctx, int64(9)).Return(nil, sql.ErrNoRows)
		_, err := b.SetBookTags(ctx, 9, []string{"a"})
		require.ErrorIs(t, err, sql.ErrNoRows)
	})
}

func TestBusiness_ListBooksByTag(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b, books := newTagBusiness(t)
	books.EXPECT().ListByTag(ctx, "sci-fi", int64(5), int64(10)).Return([]db.Book{{ID: 1}}, nil)
	got, err := b.ListBooksByTag(ctx, " Sci-Fi ", 5, 10)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	_, err = b.ListBooksByTag(ctx, "bad tag", 5, 10)
	require.ErrorIs(t, err, business.ErrInvalidTag)
}

func TestBusiness_WithTags(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	b, books := newTagBusiness(t)
	books.EXPECT().TagsFor(ctx, mock.Anything).Return(map[int64][]string{2: {"a"}}, nil)
	got, err := b.WithTags(ctx, []db.Book{{ID: 1}, {ID: 2}})
	require.NoError(t, err)
	assert.Equal(t, []string{}, got[0].Tags)
	assert.Equal(t, []string{"a"}, got[1].Tags)
}

func TestBusiness_ListTags(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("normalizes the prefix and passes the page", func(t *testing.T) {
		t.Parallel()
		b, books := newTagBusiness(t)
		want := []storage.TagCount{{Tag: "fantasy", Books: 3}}
		books.EXPECT().ListTagCounts(ctx, "fan", int64(5), int64(10)).Return(want, 7, nil)
		got, total, err := b.ListTags(ctx, "  Fan ", 5, 10)
		require.NoError(t, err)
		assert.Equal(t, want, got)
		assert.Equal(t, int64(7), total)
	})

	t.Run("empty prefix matches all and nil becomes empty", func(t *testing.T) {
		t.Parallel()
		b, books := newTagBusiness(t)
		books.EXPECT().ListTagCounts(ctx, "", int64(20), int64(0)).Return(nil, 0, nil)
		got, total, err := b.ListTags(ctx, "", 20, 0)
		require.NoError(t, err)
		assert.NotNil(t, got)
		assert.Empty(t, got)
		assert.Zero(t, total)
	})

	t.Run("invalid prefix", func(t *testing.T) {
		t.Parallel()
		b, _ := newTagBusiness(t)
		for _, p := range []string{"a b", "a_b", "café", "a/", strings.Repeat("a", 33)} {
			_, _, err := b.ListTags(ctx, p, 20, 0)
			require.ErrorIs(t, err, business.ErrInvalidTag, p)
		}
	})

	t.Run("longest valid prefix", func(t *testing.T) {
		t.Parallel()
		b, books := newTagBusiness(t)
		p := strings.Repeat("a", 32)
		books.EXPECT().ListTagCounts(ctx, p, int64(20), int64(0)).Return(nil, 0, nil)
		_, _, err := b.ListTags(ctx, p, 20, 0)
		require.NoError(t, err)
	})

	t.Run("storage error", func(t *testing.T) {
		t.Parallel()
		b, books := newTagBusiness(t)
		books.EXPECT().ListTagCounts(ctx, "", int64(20), int64(0)).Return(nil, 0, sql.ErrConnDone)
		_, _, err := b.ListTags(ctx, "", 20, 0)
		require.ErrorIs(t, err, sql.ErrConnDone)
	})
}
