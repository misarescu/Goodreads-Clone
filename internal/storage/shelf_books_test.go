package storage

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShelfBookRepository_Add(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	user := createTestUser(t, s, "test@example.com", "testuser")
	shelf := createTestShelf(t, s, user.ID, "To Read", false)
	book := createTestBook(t, s, "Test Book", "Test Author")

	t.Run("successful add", func(t *testing.T) {
		shelfBook, err := s.ShelfBooks.Add(ctx, shelf.ID, book.ID, 0, 0)
		require.NoError(t, err)
		require.NotNil(t, shelfBook)
		assert.NotZero(t, shelfBook.ID)
		assert.Equal(t, shelf.ID, shelfBook.ShelfID)
		assert.Equal(t, book.ID, shelfBook.BookID)
		assert.Equal(t, int64(0), shelfBook.PagesRead)
		assert.Equal(t, int64(0), shelfBook.ProgressPercent)
		assert.NotZero(t, shelfBook.AddedAt)
		assert.NotZero(t, shelfBook.UpdatedAt)
	})

	t.Run("add with progress", func(t *testing.T) {
		book2 := createTestBook(t, s, "Another Book", "Another Author")
		shelfBook, err := s.ShelfBooks.Add(ctx, shelf.ID, book2.ID, 150, 50)
		require.NoError(t, err)
		assert.Equal(t, int64(150), shelfBook.PagesRead)
		assert.Equal(t, int64(50), shelfBook.ProgressPercent)
	})

	t.Run("duplicate shelf_book error", func(t *testing.T) {
		shelf2 := createTestShelf(t, s, user.ID, "Currently Reading", false)
		book3 := createTestBook(t, s, "Duplicate Test", "Author")

		// Add book to shelf first time
		_, err := s.ShelfBooks.Add(ctx, shelf2.ID, book3.ID, 0, 0)
		require.NoError(t, err)

		// Attempt to add same book to same shelf again
		shelfBook, err := s.ShelfBooks.Add(ctx, shelf2.ID, book3.ID, 0, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "UNIQUE constraint failed")
		assert.Nil(t, shelfBook)
	})

	t.Run("same book on different shelves", func(t *testing.T) {
		shelf3 := createTestShelf(t, s, user.ID, "Favorites", false)
		shelf4 := createTestShelf(t, s, user.ID, "Finished", false)
		book4 := createTestBook(t, s, "Multi-Shelf Book", "Author")

		// Add same book to two different shelves
		shelfBook1, err := s.ShelfBooks.Add(ctx, shelf3.ID, book4.ID, 0, 0)
		require.NoError(t, err)

		shelfBook2, err := s.ShelfBooks.Add(ctx, shelf4.ID, book4.ID, 300, 100)
		require.NoError(t, err)

		assert.NotEqual(t, shelfBook1.ID, shelfBook2.ID)
		assert.Equal(t, book4.ID, shelfBook1.BookID)
		assert.Equal(t, book4.ID, shelfBook2.BookID)
	})
}

func TestShelfBookRepository_Get(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	user := createTestUser(t, s, "test@example.com", "testuser")
	shelf := createTestShelf(t, s, user.ID, "My Shelf", false)
	book := createTestBook(t, s, "Test Book", "Test Author")

	// Add book to shelf
	added, err := s.ShelfBooks.Add(ctx, shelf.ID, book.ID, 100, 25)
	require.NoError(t, err)

	t.Run("existing shelf_book", func(t *testing.T) {
		shelfBook, err := s.ShelfBooks.Get(ctx, shelf.ID, book.ID)
		require.NoError(t, err)
		assert.Equal(t, added.ID, shelfBook.ID)
		assert.Equal(t, shelf.ID, shelfBook.ShelfID)
		assert.Equal(t, book.ID, shelfBook.BookID)
		assert.Equal(t, int64(100), shelfBook.PagesRead)
		assert.Equal(t, int64(25), shelfBook.ProgressPercent)
	})

	t.Run("non-existent shelf_book", func(t *testing.T) {
		shelfBook, err := s.ShelfBooks.Get(ctx, 99999, book.ID)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, shelfBook)
	})
}

func TestShelfBookRepository_UpdateProgress(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	user := createTestUser(t, s, "test@example.com", "testuser")
	shelf := createTestShelf(t, s, user.ID, "Currently Reading", false)
	book := createTestBook(t, s, "In Progress Book", "Author")

	// Add book to shelf
	added, err := s.ShelfBooks.Add(ctx, shelf.ID, book.ID, 0, 0)
	require.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		updated, err := s.ShelfBooks.UpdateProgress(ctx, added.ID, 200, 50)
		require.NoError(t, err)
		assert.Equal(t, added.ID, updated.ID)
		assert.Equal(t, int64(200), updated.PagesRead)
		assert.Equal(t, int64(50), updated.ProgressPercent)
		assert.True(t, updated.UpdatedAt.After(added.UpdatedAt) || updated.UpdatedAt.Equal(added.UpdatedAt))
	})

	t.Run("update to completion", func(t *testing.T) {
		updated, err := s.ShelfBooks.UpdateProgress(ctx, added.ID, 400, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(400), updated.PagesRead)
		assert.Equal(t, int64(100), updated.ProgressPercent)
	})

	t.Run("update non-existent shelf_book", func(t *testing.T) {
		shelfBook, err := s.ShelfBooks.UpdateProgress(ctx, 99999, 100, 50)
		require.Error(t, err)
		assert.Nil(t, shelfBook)
	})
}

func TestShelfBookRepository_Remove(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	user := createTestUser(t, s, "test@example.com", "testuser")
	shelf := createTestShelf(t, s, user.ID, "My Shelf", false)
	book := createTestBook(t, s, "To Remove", "Author")

	// Add book to shelf
	_, err := s.ShelfBooks.Add(ctx, shelf.ID, book.ID, 0, 0)
	require.NoError(t, err)

	t.Run("successful remove", func(t *testing.T) {
		err = s.ShelfBooks.Remove(ctx, shelf.ID, book.ID)
		require.NoError(t, err)

		// Verify removal
		_, err = s.ShelfBooks.Get(ctx, shelf.ID, book.ID)
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("remove non-existent shelf_book", func(t *testing.T) {
		err = s.ShelfBooks.Remove(ctx, 99999, book.ID)
		assert.NoError(t, err) // SQLite doesn't error on DELETE with no rows affected
	})
}

func TestShelfBookRepository_ListBooksOnShelf(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	user := createTestUser(t, s, "test@example.com", "testuser")
	shelf := createTestShelf(t, s, user.ID, "My Collection", false)

	// Create and add multiple books
	book1 := createTestBook(t, s, "Book One", "Author A")
	book2 := createTestBook(t, s, "Book Two", "Author B")
	book3 := createTestBook(t, s, "Book Three", "Author C")

	_, err := s.ShelfBooks.Add(ctx, shelf.ID, book1.ID, 0, 0)
	require.NoError(t, err)
	_, err = s.ShelfBooks.Add(ctx, shelf.ID, book2.ID, 150, 50)
	require.NoError(t, err)
	_, err = s.ShelfBooks.Add(ctx, shelf.ID, book3.ID, 300, 100)
	require.NoError(t, err)

	t.Run("list all books on shelf", func(t *testing.T) {
		books, err := s.ShelfBooks.ListBooksOnShelf(ctx, shelf.ID)
		require.NoError(t, err)
		assert.Len(t, books, 3)

		// Verify JOIN returns book details
		titles := make(map[string]bool)
		for _, row := range books {
			titles[row.Title] = true
			assert.Equal(t, shelf.ID, row.ShelfID)
			assert.NotZero(t, row.BookID)
			assert.NotEmpty(t, row.Title)
			assert.NotEmpty(t, row.Author)
		}

		assert.True(t, titles["Book One"])
		assert.True(t, titles["Book Two"])
		assert.True(t, titles["Book Three"])
	})

	t.Run("empty shelf", func(t *testing.T) {
		emptyShelf := createTestShelf(t, s, user.ID, "Empty Shelf", false)
		books, err := s.ShelfBooks.ListBooksOnShelf(ctx, emptyShelf.ID)
		require.NoError(t, err)
		assert.Empty(t, books)
	})

	t.Run("verify progress tracking", func(t *testing.T) {
		books, err := s.ShelfBooks.ListBooksOnShelf(ctx, shelf.ID)
		require.NoError(t, err)

		// Find book2 which has progress
		for _, row := range books {
			if row.Title == "Book Two" {
				assert.Equal(t, int64(150), row.PagesRead)
				assert.Equal(t, int64(50), row.ProgressPercent)
			}
		}
	})
}

func TestShelfBookRepository_GetUserBookShelves(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	user := createTestUser(t, s, "test@example.com", "testuser")
	book := createTestBook(t, s, "Popular Book", "Famous Author")

	// Create multiple shelves
	shelf1 := createTestShelf(t, s, user.ID, "To Read", true)
	shelf2 := createTestShelf(t, s, user.ID, "Favorites", false)
	shelf3 := createTestShelf(t, s, user.ID, "Finished", true)
	_ = createTestShelf(t, s, user.ID, "DNF", true) // Create but don't add book to this shelf

	// Add book to multiple shelves
	_, err := s.ShelfBooks.Add(ctx, shelf1.ID, book.ID, 0, 0)
	require.NoError(t, err)
	_, err = s.ShelfBooks.Add(ctx, shelf2.ID, book.ID, 0, 0)
	require.NoError(t, err)
	_, err = s.ShelfBooks.Add(ctx, shelf3.ID, book.ID, 400, 100)
	require.NoError(t, err)

	t.Run("book on multiple shelves", func(t *testing.T) {
		shelves, err := s.ShelfBooks.GetUserBookShelves(ctx, user.ID, book.ID)
		require.NoError(t, err)
		assert.Len(t, shelves, 3)

		// Verify shelf names
		names := make(map[string]bool)
		for _, shelf := range shelves {
			names[shelf.Name] = true
			assert.Equal(t, user.ID, shelf.UserID)
		}

		assert.True(t, names["To Read"])
		assert.True(t, names["Favorites"])
		assert.True(t, names["Finished"])
		assert.False(t, names["DNF"]) // Book not on this shelf
	})

	t.Run("book not on any shelves", func(t *testing.T) {
		book2 := createTestBook(t, s, "Unlisted Book", "Author")
		shelves, err := s.ShelfBooks.GetUserBookShelves(ctx, user.ID, book2.ID)
		require.NoError(t, err)
		assert.Empty(t, shelves)
	})

	t.Run("different user", func(t *testing.T) {
		user2 := createTestUser(t, s, "user2@example.com", "user2")
		shelves, err := s.ShelfBooks.GetUserBookShelves(ctx, user2.ID, book.ID)
		require.NoError(t, err)
		assert.Empty(t, shelves)
	})
}

func TestShelfBookRepository_ForeignKeyCascades(t *testing.T) {
	t.Parallel()

	t.Run("cascade on shelf delete", func(t *testing.T) {
		s := newTestStorage(t)
		ctx := context.Background()

		user := createTestUser(t, s, "test@example.com", "testuser")
		shelf := createTestShelf(t, s, user.ID, "Temporary Shelf", false)
		book := createTestBook(t, s, "Test Book", "Author")

		_, err := s.ShelfBooks.Add(ctx, shelf.ID, book.ID, 0, 0)
		require.NoError(t, err)

		// Delete shelf
		err = s.Shelves.Delete(ctx, shelf.ID)
		require.NoError(t, err)

		// Verify shelf_books entry was cascade deleted
		_, err = s.ShelfBooks.Get(ctx, shelf.ID, book.ID)
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("cascade on book delete", func(t *testing.T) {
		s := newTestStorage(t)
		ctx := context.Background()

		user := createTestUser(t, s, "test@example.com", "testuser")
		shelf := createTestShelf(t, s, user.ID, "My Shelf", false)
		book := createTestBook(t, s, "Temporary Book", "Author")

		_, err := s.ShelfBooks.Add(ctx, shelf.ID, book.ID, 0, 0)
		require.NoError(t, err)

		// Delete book
		err = s.Books.Delete(ctx, book.ID)
		require.NoError(t, err)

		// Verify shelf_books entry was cascade deleted
		_, err = s.ShelfBooks.Get(ctx, shelf.ID, book.ID)
		require.ErrorIs(t, err, sql.ErrNoRows)
	})
}
