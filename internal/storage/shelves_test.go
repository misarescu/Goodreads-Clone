package storage

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShelfRepository_Create(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	user := createTestUser(t, s, "test@example.com", "testuser")

	t.Run("successful creation with is_default=true", func(t *testing.T) {
		shelf, err := s.Shelves.Create(ctx, user.ID, "To Read", true)
		require.NoError(t, err)
		require.NotNil(t, shelf)
		assert.NotZero(t, shelf.ID)
		assert.Equal(t, user.ID, shelf.UserID)
		assert.Equal(t, "To Read", shelf.Name)
		assert.True(t, shelf.IsDefault)
		assert.NotZero(t, shelf.CreatedAt)
	})

	t.Run("successful creation with is_default=false", func(t *testing.T) {
		shelf, err := s.Shelves.Create(ctx, user.ID, "Favorites", false)
		require.NoError(t, err)
		assert.Equal(t, "Favorites", shelf.Name)
		assert.False(t, shelf.IsDefault)
	})

	t.Run("duplicate name for same user error", func(t *testing.T) {
		// Create first shelf
		_, err := s.Shelves.Create(ctx, user.ID, "Duplicated", false)
		require.NoError(t, err)

		// Attempt to create shelf with same name for same user
		shelf, err := s.Shelves.Create(ctx, user.ID, "Duplicated", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "UNIQUE constraint failed")
		assert.Nil(t, shelf)
	})

	t.Run("different users can have same shelf name", func(t *testing.T) {
		user2 := createTestUser(t, s, "user2@example.com", "user2")

		// Create shelf with same name for different users
		shelf1, err := s.Shelves.Create(ctx, user.ID, "SharedName", false)
		require.NoError(t, err)

		shelf2, err := s.Shelves.Create(ctx, user2.ID, "SharedName", false)
		require.NoError(t, err)

		assert.NotEqual(t, shelf1.ID, shelf2.ID)
		assert.Equal(t, shelf1.Name, shelf2.Name)
		assert.NotEqual(t, shelf1.UserID, shelf2.UserID)
	})
}

func TestShelfRepository_GetByID(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	user := createTestUser(t, s, "test@example.com", "testuser")
	created := createTestShelf(t, s, user.ID, "Test Shelf", false)

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "existing shelf",
			id:      created.ID,
			wantErr: false,
		},
		{
			name:    "non-existent shelf",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shelf, err := s.Shelves.GetByID(ctx, tt.id)

			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, sql.ErrNoRows)
			} else {
				require.NoError(t, err)
				assert.Equal(t, created.ID, shelf.ID)
				assert.Equal(t, created.Name, shelf.Name)
				assert.Equal(t, created.UserID, shelf.UserID)
			}
		})
	}
}

func TestShelfRepository_GetByUserAndName(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	user := createTestUser(t, s, "test@example.com", "testuser")
	created := createTestShelf(t, s, user.ID, "My Shelf", false)

	t.Run("existing shelf", func(t *testing.T) {
		shelf, err := s.Shelves.GetByUserAndName(ctx, user.ID, "My Shelf")
		require.NoError(t, err)
		assert.Equal(t, created.ID, shelf.ID)
		assert.Equal(t, user.ID, shelf.UserID)
		assert.Equal(t, "My Shelf", shelf.Name)
	})

	t.Run("non-existent shelf", func(t *testing.T) {
		shelf, err := s.Shelves.GetByUserAndName(ctx, user.ID, "Nonexistent")
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, shelf)
	})

	t.Run("wrong user", func(t *testing.T) {
		shelf, err := s.Shelves.GetByUserAndName(ctx, 99999, "My Shelf")
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, shelf)
	})
}

func TestShelfRepository_ListByUser(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	user1 := createTestUser(t, s, "user1@example.com", "user1")
	user2 := createTestUser(t, s, "user2@example.com", "user2")

	// Create shelves for user1
	createTestShelf(t, s, user1.ID, "Shelf 1", false)
	createTestShelf(t, s, user1.ID, "Shelf 2", false)
	createTestShelf(t, s, user1.ID, "Shelf 3", true)

	// Create shelf for user2
	createTestShelf(t, s, user2.ID, "Other User Shelf", false)

	t.Run("user1 shelves", func(t *testing.T) {
		shelves, err := s.Shelves.ListByUser(ctx, user1.ID)
		require.NoError(t, err)
		assert.Len(t, shelves, 3)

		// Verify all shelves belong to user1
		for _, shelf := range shelves {
			assert.Equal(t, user1.ID, shelf.UserID)
		}
	})

	t.Run("user2 shelves", func(t *testing.T) {
		shelves, err := s.Shelves.ListByUser(ctx, user2.ID)
		require.NoError(t, err)
		assert.Len(t, shelves, 1)
		assert.Equal(t, "Other User Shelf", shelves[0].Name)
	})

	t.Run("user with no shelves", func(t *testing.T) {
		user3 := createTestUser(t, s, "user3@example.com", "user3")
		shelves, err := s.Shelves.ListByUser(ctx, user3.ID)
		require.NoError(t, err)
		assert.Empty(t, shelves)
	})
}

func TestShelfRepository_Update(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	user := createTestUser(t, s, "test@example.com", "testuser")
	created := createTestShelf(t, s, user.ID, "Original Name", false)

	t.Run("successful update", func(t *testing.T) {
		updated, err := s.Shelves.Update(ctx, created.ID, "Updated Name")
		require.NoError(t, err)
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, created.UserID, updated.UserID)
	})

	t.Run("update non-existent shelf", func(t *testing.T) {
		shelf, err := s.Shelves.Update(ctx, 99999, "Should Fail")
		require.Error(t, err)
		assert.Nil(t, shelf)
	})
}

func TestShelfRepository_Delete(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	user := createTestUser(t, s, "test@example.com", "testuser")
	created := createTestShelf(t, s, user.ID, "To Delete", false)

	t.Run("successful delete", func(t *testing.T) {
		err := s.Shelves.Delete(ctx, created.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = s.Shelves.GetByID(ctx, created.ID)
		require.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("delete non-existent shelf", func(t *testing.T) {
		err := s.Shelves.Delete(ctx, 99999)
		assert.NoError(t, err) // SQLite doesn't error on DELETE with no rows affected
	})
}

func TestShelfRepository_GetDefaultsByUser(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	user := createTestUser(t, s, "test@example.com", "testuser")

	// Create mix of default and custom shelves
	createTestShelf(t, s, user.ID, "To Read", true)
	createTestShelf(t, s, user.ID, "Currently Reading", true)
	createTestShelf(t, s, user.ID, "Custom Shelf", false)
	createTestShelf(t, s, user.ID, "Another Custom", false)

	t.Run("only default shelves returned", func(t *testing.T) {
		shelves, err := s.Shelves.GetDefaultsByUser(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, shelves, 2)

		// Verify all returned shelves are default
		for _, shelf := range shelves {
			assert.True(t, shelf.IsDefault)
			assert.Equal(t, user.ID, shelf.UserID)
		}
	})

	t.Run("user with no default shelves", func(t *testing.T) {
		user2 := createTestUser(t, s, "user2@example.com", "user2")
		createTestShelf(t, s, user2.ID, "Only Custom", false)

		shelves, err := s.Shelves.GetDefaultsByUser(ctx, user2.ID)
		require.NoError(t, err)
		assert.Empty(t, shelves)
	})
}

func TestShelfRepository_Delete_CascadeShelfBooks(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	// Create user, shelf, and book
	user := createTestUser(t, s, "test@example.com", "testuser")
	shelf := createTestShelf(t, s, user.ID, "My Shelf", false)
	book := createTestBook(t, s, "Test Book", "Test Author")

	// Add book to shelf
	_, err := s.ShelfBooks.Add(ctx, shelf.ID, book.ID, 0, 0)
	require.NoError(t, err)

	// Delete the shelf
	err = s.Shelves.Delete(ctx, shelf.ID)
	require.NoError(t, err)

	// Verify shelf_books entry was cascade deleted
	_, err = s.ShelfBooks.Get(ctx, shelf.ID, book.ID)
	require.ErrorIs(t, err, sql.ErrNoRows)
}
