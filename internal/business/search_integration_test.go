package business_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"goodreads-clone.com/internal/business"
	"goodreads-clone.com/internal/db"
	"goodreads-clone.com/internal/storage"
)

func lang(l string) sql.NullString { return sql.NullString{String: l, Valid: l != ""} }

func seedSearchBooks(t *testing.T, s *storage.Storage) {
	t.Helper()
	for _, b := range []db.CreateBookParams{
		{Title: "The Hobbit", Author: "J. R. R. Tolkien", Language: lang("en")},
		{Title: "The Fellowship of the Ring", Author: "J. R. R. Tolkien", Language: lang("en")},
		{Title: "Der Hobbit", Author: "J. R. R. Tolkien", Language: lang("de")},
		{Title: "Tolkien: A Biography", Author: "Humphrey Carpenter", Language: lang("en")},
		{Title: "Dune", Author: "Frank Herbert", Language: lang("en")},
	} {
		_, err := s.Books.Create(context.Background(), b)
		require.NoError(t, err)
	}
}

func search(t *testing.T, s *storage.Storage, q string, limit, offset int64) ([]string, int64) {
	t.Helper()
	pq, err := business.ParseSearchQuery(q)
	require.NoError(t, err, q)
	books, total, err := s.Books.Search(context.Background(), storage.SearchParams{
		Match: pq.Match, Language: pq.Language, Limit: limit, Offset: offset,
	})
	require.NoError(t, err, q)
	titles := make([]string, 0, len(books))
	for _, b := range books {
		titles = append(titles, b.Title)
	}
	return titles, total
}

func TestBookRepository_Search(t *testing.T) {
	t.Parallel()
	s := newSearchStorage(t)
	seedSearchBooks(t, s)

	t.Run("prefix matching on the last word", func(t *testing.T) {
		titles, total := search(t, s, "tolk", 20, 0)
		assert.Equal(t, int64(4), total)
		assert.Len(t, titles, 4)
	})

	t.Run("a non-final word is not a prefix", func(t *testing.T) {
		titles, _ := search(t, s, "tolk hobbit", 20, 0)
		assert.Empty(t, titles)
	})

	t.Run("title match ranks above author match", func(t *testing.T) {
		titles, _ := search(t, s, "tolkien", 20, 0)
		require.Len(t, titles, 4)
		assert.Equal(t, "Tolkien: A Biography", titles[0])
	})

	t.Run("author filter", func(t *testing.T) {
		titles, total := search(t, s, "hobbit author:tolkien", 20, 0)
		assert.Equal(t, int64(2), total)
		assert.ElementsMatch(t, []string{"The Hobbit", "Der Hobbit"}, titles)

		titles, _ = search(t, s, "author:carpenter", 20, 0)
		assert.Equal(t, []string{"Tolkien: A Biography"}, titles)

		// the filter is limited to the author column
		titles, _ = search(t, s, "author:hobbit", 20, 0)
		assert.Empty(t, titles)
	})

	t.Run("lang filter", func(t *testing.T) {
		titles, total := search(t, s, "hobbit lang:de", 20, 0)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, []string{"Der Hobbit"}, titles)

		titles, _ = search(t, s, "hobbit author:tolkien lang:EN", 20, 0)
		assert.Equal(t, []string{"The Hobbit"}, titles)

		titles, _ = search(t, s, "hobbit lang:fr", 20, 0)
		assert.Empty(t, titles)
	})

	t.Run("pagination", func(t *testing.T) {
		all, total := search(t, s, "tolk", 20, 0)
		page1, total1 := search(t, s, "tolk", 3, 0)
		page2, total2 := search(t, s, "tolk", 3, 3)
		assert.Equal(t, int64(4), total)
		assert.Equal(t, total, total1)
		assert.Equal(t, total, total2)
		assert.Equal(t, all, append(append([]string{}, page1...), page2...))
		assert.Len(t, page1, 3)
		assert.Len(t, page2, 1)

		beyond, totalBeyond := search(t, s, "tolk", 3, 100)
		assert.Empty(t, beyond)
		assert.Equal(t, total, totalBeyond)
	})

	t.Run("hostile input never errors", func(t *testing.T) {
		for _, in := range []string{
			`"`, `""`, `"unterminated`, `hobbit"`, `(`, `)`, `*`, `a*b*`, `^hobbit`, `title:hobbit`,
			`hobbit AND`, `OR`, `NOT hobbit`, `NEAR(hobbit dune)`, `{title}:hobbit`, `-hobbit`,
			`hobbit'; DROP TABLE books;--`, "author:\"x", `author:(a OR b)`, `lang:'`, "\x00hobbit",
			`books_fts`, `rank`, `+ - * : ^ " ( ) { }`,
		} {
			pq, err := business.ParseSearchQuery(in)
			if err != nil {
				require.ErrorIs(t, err, business.ErrInvalidSearch, in)
				continue
			}
			_, _, err = s.Books.Search(context.Background(), storage.SearchParams{
				Match: pq.Match, Language: pq.Language, Limit: 10,
			})
			require.NoError(t, err, "input %q -> %q", in, pq.Match)
		}
		_, err := s.Books.GetByID(context.Background(), 1)
		assert.NoError(t, err, "books table must be intact")
	})
}

func TestBookRepository_SearchStaysInSync(t *testing.T) {
	t.Parallel()
	s := newSearchStorage(t)
	ctx := context.Background()

	book, err := s.Books.Create(ctx, db.CreateBookParams{Title: "Old Title", Author: "Ann Author"})
	require.NoError(t, err)
	titles, _ := search(t, s, "old", 10, 0)
	assert.Equal(t, []string{"Old Title"}, titles)

	_, err = s.Books.Update(ctx, db.UpdateBookParams{ID: book.ID, Title: "New Title", Author: "Bob Writer"})
	require.NoError(t, err)
	titles, _ = search(t, s, "old", 10, 0)
	assert.Empty(t, titles)
	titles, _ = search(t, s, "new author:bob", 10, 0)
	assert.Equal(t, []string{"New Title"}, titles)
	titles, _ = search(t, s, "ann", 10, 0)
	assert.Empty(t, titles)

	require.NoError(t, s.Books.Delete(ctx, book.ID))
	titles, total := search(t, s, "new", 10, 0)
	assert.Empty(t, titles)
	assert.Zero(t, total)
}

func TestBooksFTSMigrationBackfillsExistingRows(t *testing.T) {
	t.Parallel()
	database, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })
	database.SetMaxOpenConns(1)

	_, err = database.Exec(`CREATE TABLE books (id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL, author TEXT NOT NULL, language TEXT);
INSERT INTO books (title, author) VALUES ('Dune', 'Frank Herbert');`)
	require.NoError(t, err)
	migration, err := readMigration("006_books_fts.sql")
	require.NoError(t, err)
	_, err = database.Exec(migration)
	require.NoError(t, err)
	// Re-applying must be harmless (make migrate re-runs every file).
	_, err = database.Exec(migration)
	require.NoError(t, err)

	var n int
	require.NoError(t, database.QueryRow(`SELECT COUNT(*) FROM books_fts('herbert')`).Scan(&n))
	assert.Equal(t, 1, n)
}

func readMigration(name string) (string, error) {
	b, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", name))
	return string(b), err
}

// newSearchStorage returns storage over a real in-memory SQLite database with all migrations applied.
func newSearchStorage(t *testing.T) *storage.Storage {
	t.Helper()
	database, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	// One connection: every connection to ":memory:" is its own database.
	database.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = database.Close() })

	files, err := filepath.Glob(filepath.Join("..", "..", "db", "migrations", "*.sql"))
	require.NoError(t, err)
	require.NotEmpty(t, files)
	sort.Strings(files)
	for _, f := range files {
		sqlText, err := os.ReadFile(f)
		require.NoError(t, err)
		_, err = database.Exec(string(sqlText))
		require.NoError(t, err, f)
	}
	return storage.New(database)
}
