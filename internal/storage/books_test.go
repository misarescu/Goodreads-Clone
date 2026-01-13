package storage

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goodreads-clone.com/internal/db"
)

func TestBookRepository_Create(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	t.Run("successful creation with all fields", func(t *testing.T) {
		releaseDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		params := db.CreateBookParams{
			Title:  "The Great Book",
			Author: "John Doe",
			PageCount: sql.NullInt64{
				Int64: 350,
				Valid: true,
			},
			Isbn: sql.NullString{
				String: "978-0123456789",
				Valid:  true,
			},
			ReleaseDate: sql.NullTime{
				Time:  releaseDate,
				Valid: true,
			},
			Language: sql.NullString{
				String: "English",
				Valid:  true,
			},
		}

		book, err := s.Books.Create(ctx, params)
		require.NoError(t, err)
		require.NotNil(t, book)
		assert.NotZero(t, book.ID)
		assert.Equal(t, "The Great Book", book.Title)
		assert.Equal(t, "John Doe", book.Author)
		assert.True(t, book.PageCount.Valid)
		assert.Equal(t, int64(350), book.PageCount.Int64)
		assert.True(t, book.Isbn.Valid)
		assert.Equal(t, "978-0123456789", book.Isbn.String)
		assert.True(t, book.ReleaseDate.Valid)
		assert.True(t, book.Language.Valid)
		assert.Equal(t, "English", book.Language.String)
		assert.NotZero(t, book.CreatedAt)
	})

	t.Run("successful creation with only required fields", func(t *testing.T) {
		params := db.CreateBookParams{
			Title:  "Minimal Book",
			Author: "Jane Smith",
		}

		book, err := s.Books.Create(ctx, params)
		require.NoError(t, err)
		require.NotNil(t, book)
		assert.Equal(t, "Minimal Book", book.Title)
		assert.Equal(t, "Jane Smith", book.Author)
		assert.False(t, book.PageCount.Valid)
		assert.False(t, book.Isbn.Valid)
		assert.False(t, book.ReleaseDate.Valid)
		assert.False(t, book.Language.Valid)
	})

	t.Run("duplicate ISBN error", func(t *testing.T) {
		isbn := "978-1234567890"
		params1 := db.CreateBookParams{
			Title:  "First Book",
			Author: "Author One",
			Isbn: sql.NullString{
				String: isbn,
				Valid:  true,
			},
		}

		_, err := s.Books.Create(ctx, params1)
		require.NoError(t, err)

		// Attempt to create another book with the same ISBN
		params2 := db.CreateBookParams{
			Title:  "Second Book",
			Author: "Author Two",
			Isbn: sql.NullString{
				String: isbn,
				Valid:  true,
			},
		}

		book, err := s.Books.Create(ctx, params2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "UNIQUE constraint failed")
		assert.Nil(t, book)
	})
}

func TestBookRepository_GetByID(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	created := createTestBook(t, s, "Test Book", "Test Author")

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "existing book",
			id:      created.ID,
			wantErr: false,
		},
		{
			name:    "non-existent book",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			book, err := s.Books.GetByID(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, sql.ErrNoRows)
			} else {
				require.NoError(t, err)
				assert.Equal(t, created.ID, book.ID)
				assert.Equal(t, created.Title, book.Title)
				assert.Equal(t, created.Author, book.Author)
			}
		})
	}
}

func TestBookRepository_GetByISBN(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	isbn := "978-9876543210"
	params := db.CreateBookParams{
		Title:  "ISBN Book",
		Author: "ISBN Author",
		Isbn: sql.NullString{
			String: isbn,
			Valid:  true,
		},
	}
	created, err := s.Books.Create(ctx, params)
	require.NoError(t, err)

	t.Run("existing ISBN", func(t *testing.T) {
		book, err := s.Books.GetByISBN(ctx, isbn)
		require.NoError(t, err)
		assert.Equal(t, created.ID, book.ID)
		assert.Equal(t, isbn, book.Isbn.String)
	})

	t.Run("non-existent ISBN", func(t *testing.T) {
		book, err := s.Books.GetByISBN(ctx, "978-0000000000")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, book)
	})
}

func TestBookRepository_Update(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	created := createTestBook(t, s, "Original Title", "Original Author")

	t.Run("successful update", func(t *testing.T) {
		params := db.UpdateBookParams{
			ID:     created.ID,
			Title:  "Updated Title",
			Author: "Updated Author",
			PageCount: sql.NullInt64{
				Int64: 500,
				Valid: true,
			},
			Isbn: sql.NullString{
				String: "978-1111111111",
				Valid:  true,
			},
			Language: sql.NullString{
				String: "French",
				Valid:  true,
			},
		}

		updated, err := s.Books.Update(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, "Updated Title", updated.Title)
		assert.Equal(t, "Updated Author", updated.Author)
		assert.True(t, updated.PageCount.Valid)
		assert.Equal(t, int64(500), updated.PageCount.Int64)
		assert.True(t, updated.Isbn.Valid)
		assert.Equal(t, "978-1111111111", updated.Isbn.String)
	})

	t.Run("update non-existent book", func(t *testing.T) {
		params := db.UpdateBookParams{
			ID:     99999,
			Title:  "Should Fail",
			Author: "Should Fail",
		}

		book, err := s.Books.Update(ctx, params)
		assert.Error(t, err)
		assert.Nil(t, book)
	})
}

func TestBookRepository_Delete(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	created := createTestBook(t, s, "To Delete", "Author")

	t.Run("successful delete", func(t *testing.T) {
		err := s.Books.Delete(ctx, created.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = s.Books.GetByID(ctx, created.ID)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("delete non-existent book", func(t *testing.T) {
		err := s.Books.Delete(ctx, 99999)
		assert.NoError(t, err) // SQLite doesn't error on DELETE with no rows affected
	})
}

func TestBookRepository_List(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	// Create multiple books
	for i := 0; i < 5; i++ {
		title := fmt.Sprintf("Book %d", i)
		author := fmt.Sprintf("Author %d", i)
		_, err := s.Books.Create(ctx, db.CreateBookParams{
			Title:  title,
			Author: author,
		})
		require.NoError(t, err)
	}

	tests := []struct {
		name      string
		limit     int64
		offset    int64
		wantCount int
	}{
		{
			name:      "first page",
			limit:     2,
			offset:    0,
			wantCount: 2,
		},
		{
			name:      "second page",
			limit:     2,
			offset:    2,
			wantCount: 2,
		},
		{
			name:      "last page",
			limit:     2,
			offset:    4,
			wantCount: 1,
		},
		{
			name:      "all books",
			limit:     10,
			offset:    0,
			wantCount: 5,
		},
		{
			name:      "offset beyond results",
			limit:     10,
			offset:    100,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			books, err := s.Books.List(ctx, tt.limit, tt.offset)
			require.NoError(t, err)
			assert.Len(t, books, tt.wantCount)
		})
	}
}

func TestBookRepository_SearchByTitle(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	// Create test books
	testBooks := []struct {
		title  string
		author string
	}{
		{"The Great Gatsby", "F. Scott Fitzgerald"},
		{"Great Expectations", "Charles Dickens"},
		{"To Kill a Mockingbird", "Harper Lee"},
		{"1984", "George Orwell"},
	}

	for _, tb := range testBooks {
		_, err := s.Books.Create(ctx, db.CreateBookParams{
			Title:  tb.title,
			Author: tb.author,
		})
		require.NoError(t, err)
	}

	tests := []struct {
		name        string
		query       string
		limit       int64
		minResults  int
		checkTitles []string
	}{
		{
			name:        "exact match",
			query:       "1984",
			limit:       10,
			minResults:  1,
			checkTitles: []string{"1984"},
		},
		{
			name:       "partial match - 'Great'",
			query:      "Great",
			limit:      10,
			minResults: 2,
			checkTitles: []string{"The Great Gatsby", "Great Expectations"},
		},
		{
			name:       "case insensitive search",
			query:      "great",
			limit:      10,
			minResults: 2,
		},
		{
			name:       "no results",
			query:      "NonexistentBook",
			limit:      10,
			minResults: 0,
		},
		{
			name:       "limit results",
			query:      "Great",
			limit:      1,
			minResults: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			books, err := s.Books.SearchByTitle(ctx, tt.query, tt.limit)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(books), tt.minResults)

			// Check specific titles if provided
			if len(tt.checkTitles) > 0 {
				foundTitles := make(map[string]bool)
				for _, book := range books {
					foundTitles[book.Title] = true
				}

				for _, expectedTitle := range tt.checkTitles {
					assert.True(t, foundTitles[expectedTitle], "Expected to find title: %s", expectedTitle)
				}
			}
		})
	}
}

func TestBookRepository_Delete_CascadeShelfBooks(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	// Create user, shelf, and book
	user := createTestUser(t, s, "test@example.com", "testuser")
	shelf := createTestShelf(t, s, user.ID, "My Books", false)
	book := createTestBook(t, s, "Test Book", "Test Author")

	// Add book to shelf
	_, err := s.ShelfBooks.Add(ctx, shelf.ID, book.ID, 0, 0)
	require.NoError(t, err)

	// Delete the book
	err = s.Books.Delete(ctx, book.ID)
	require.NoError(t, err)

	// Verify shelf_books entry was cascade deleted
	_, err = s.ShelfBooks.Get(ctx, shelf.ID, book.ID)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}
