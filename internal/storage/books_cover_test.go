package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goodreads-clone.com/internal/db"
)

func TestBookRepository_CoverURL(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()
	cover := "https://example.com/covers/dune.jpg"
	isbn := "978-0-441-17271-9"

	created, err := s.Books.Create(ctx, db.CreateBookParams{Title: "Dune", Author: "Herbert", CoverUrl: &cover})
	require.NoError(t, err)
	require.NotNil(t, created.CoverUrl)
	assert.Equal(t, cover, *created.CoverUrl)

	byID, err := s.Books.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, byID.CoverUrl)
	assert.Equal(t, cover, *byID.CoverUrl)

	list, err := s.Books.List(ctx, 10, 0)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.NotNil(t, list[0].CoverUrl)

	found, total, err := s.Books.Search(ctx, SearchParams{Match: `"dune"*`, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.NotNil(t, found[0].CoverUrl)
	assert.Equal(t, cover, *found[0].CoverUrl)

	t.Run("by ISBN", func(t *testing.T) {
		withISBN, err := s.Books.Create(ctx, db.CreateBookParams{
			Title: "Emma", Author: "Austen", CoverUrl: &cover, Isbn: sql.NullString{String: isbn, Valid: true},
		})
		require.NoError(t, err)
		got, err := s.Books.GetByISBN(ctx, isbn)
		require.NoError(t, err)
		assert.Equal(t, withISBN.ID, got.ID)
		require.NotNil(t, got.CoverUrl)
	})

	t.Run("update replaces and clears", func(t *testing.T) {
		other := "http://example.com/new.png"
		updated, err := s.Books.Update(ctx, db.UpdateBookParams{
			ID: created.ID, Title: "Dune", Author: "Herbert", CoverUrl: &other,
		})
		require.NoError(t, err)
		require.NotNil(t, updated.CoverUrl)
		assert.Equal(t, other, *updated.CoverUrl)

		cleared, err := s.Books.Update(ctx, db.UpdateBookParams{ID: created.ID, Title: "Dune", Author: "Herbert"})
		require.NoError(t, err)
		assert.Nil(t, cleared.CoverUrl)
	})

	t.Run("book without a cover", func(t *testing.T) {
		b := createTestBook(t, s, "Coverless Zebra", "Nobody")
		assert.Nil(t, b.CoverUrl)
		got, err := s.Books.GetByID(ctx, b.ID)
		require.NoError(t, err)
		assert.Nil(t, got.CoverUrl)
		found, _, err := s.Books.Search(ctx, SearchParams{Match: `"zebra"*`, Limit: 10})
		require.NoError(t, err)
		require.Len(t, found, 1)
		assert.Nil(t, found[0].CoverUrl)

		out, err := json.Marshal(got)
		require.NoError(t, err)
		assert.Contains(t, string(out), `"cover_url":null`)
	})
}
