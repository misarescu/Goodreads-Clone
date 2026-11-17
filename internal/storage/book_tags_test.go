package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookRepository_Tags(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()
	a := createTestBook(t, s, "A", "x")
	b := createTestBook(t, s, "B", "x")

	require.NoError(t, s.Books.SetTags(ctx, a.ID, []string{"zeta", "alpha"}))
	require.NoError(t, s.Books.SetTags(ctx, b.ID, []string{"alpha"}))

	got, err := s.Books.TagsFor(ctx, []int64{a.ID, b.ID})
	require.NoError(t, err)
	assert.Equal(t, map[int64][]string{a.ID: {"alpha", "zeta"}, b.ID: {"alpha"}}, got, "sorted")

	t.Run("replace", func(t *testing.T) {
		require.NoError(t, s.Books.SetTags(ctx, a.ID, []string{"beta"}))
		got, err := s.Books.TagsFor(ctx, []int64{a.ID})
		require.NoError(t, err)
		assert.Equal(t, []string{"beta"}, got[a.ID])
	})

	t.Run("clear", func(t *testing.T) {
		require.NoError(t, s.Books.SetTags(ctx, b.ID, nil))
		got, err := s.Books.TagsFor(ctx, []int64{b.ID})
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("no ids", func(t *testing.T) {
		got, err := s.Books.TagsFor(ctx, nil)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestBookRepository_ListByTag(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()
	ids := make([]int64, 0, 4)
	for _, title := range []string{"A", "B", "C", "D"} {
		ids = append(ids, createTestBook(t, s, title, "x").ID)
	}
	for _, id := range ids[:3] {
		require.NoError(t, s.Books.SetTags(ctx, id, []string{"fantasy"}))
	}
	require.NoError(t, s.Books.SetTags(ctx, ids[3], []string{"fantasyx"}))

	all, err := s.Books.ListByTag(ctx, "fantasy", 10, 0)
	require.NoError(t, err)
	require.Len(t, all, 3)

	page1, err := s.Books.ListByTag(ctx, "fantasy", 2, 0)
	require.NoError(t, err)
	page2, err := s.Books.ListByTag(ctx, "fantasy", 2, 2)
	require.NoError(t, err)
	require.Len(t, page1, 2)
	require.Len(t, page2, 1)
	assert.Equal(t, all[0].ID, page1[0].ID)
	assert.Equal(t, all[2].ID, page2[0].ID)

	none, err := s.Books.ListByTag(ctx, "missing", 10, 0)
	require.NoError(t, err)
	assert.Empty(t, none)
}

func TestBookRepository_Delete_CascadeTags(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()
	book := createTestBook(t, s, "A", "x")
	require.NoError(t, s.Books.SetTags(ctx, book.ID, []string{"a", "b"}))

	require.NoError(t, s.Books.Delete(ctx, book.ID))

	var n int
	require.NoError(t, s.DB().QueryRowContext(ctx, "SELECT COUNT(*) FROM book_tags").Scan(&n))
	assert.Zero(t, n)
}

func TestBookRepository_ListTagCounts(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()
	a := createTestBook(t, s, "A", "x")
	b := createTestBook(t, s, "B", "x")
	require.NoError(t, s.Books.SetTags(ctx, a.ID, []string{"fantasy", "fan-fic", "zeta"}))
	require.NoError(t, s.Books.SetTags(ctx, b.ID, []string{"fantasy", "zeta"}))

	got, total, err := s.Books.ListTagCounts(ctx, "", 10, 0)
	require.NoError(t, err)
	assert.EqualValues(t, 3, total)
	assert.Equal(t, []TagCount{{"fantasy", 2}, {"zeta", 2}, {"fan-fic", 1}}, got)

	got, total, err = s.Books.ListTagCounts(ctx, "fan", 1, 1)
	require.NoError(t, err)
	assert.EqualValues(t, 2, total)
	assert.Equal(t, []TagCount{{"fan-fic", 1}}, got)

	require.NoError(t, s.Books.Delete(ctx, b.ID))
	got, _, err = s.Books.ListTagCounts(ctx, "", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, []TagCount{{"fan-fic", 1}, {"fantasy", 1}, {"zeta", 1}}, got)

	got, total, err = s.Books.ListTagCounts(ctx, "nope", 10, 0)
	require.NoError(t, err)
	assert.Empty(t, got)
	assert.Zero(t, total)
}
