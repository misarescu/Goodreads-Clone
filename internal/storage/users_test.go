package storage

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	tests := []struct {
		name         string
		email        string
		username     string
		passwordHash string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "successful creation",
			email:        "test@example.com",
			username:     "testuser",
			passwordHash: "hashed_password",
			wantErr:      false,
		},
		{
			name:         "another successful creation",
			email:        "user2@example.com",
			username:     "user2",
			passwordHash: "hashed_password2",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := s.Users.Create(ctx, tt.email, tt.username, tt.passwordHash)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.NotZero(t, user.ID)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.username, user.Username)
				assert.Equal(t, tt.passwordHash, user.PasswordHash)
				assert.NotZero(t, user.CreatedAt)
				assert.NotZero(t, user.UpdatedAt)
			}
		})
	}
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	// Create first user
	_, err := s.Users.Create(ctx, "test@example.com", "testuser", "hash")
	require.NoError(t, err)

	// Attempt to create user with duplicate email
	user, err := s.Users.Create(ctx, "test@example.com", "different", "hash")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UNIQUE constraint failed")
	assert.Nil(t, user)
}

func TestUserRepository_Create_DuplicateUsername(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	// Create first user
	_, err := s.Users.Create(ctx, "test@example.com", "testuser", "hash")
	require.NoError(t, err)

	// Attempt to create user with duplicate username
	user, err := s.Users.Create(ctx, "other@example.com", "testuser", "hash")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UNIQUE constraint failed")
	assert.Nil(t, user)
}

func TestUserRepository_GetByID(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	// Create a user first
	created, err := s.Users.Create(ctx, "test@example.com", "testuser", "hash")
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "existing user",
			id:      created.ID,
			wantErr: false,
		},
		{
			name:    "non-existent user",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := s.Users.GetByID(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, sql.ErrNoRows)
			} else {
				require.NoError(t, err)
				assert.Equal(t, created.ID, user.ID)
				assert.Equal(t, created.Email, user.Email)
				assert.Equal(t, created.Username, user.Username)
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	created, err := s.Users.Create(ctx, "test@example.com", "testuser", "hash")
	require.NoError(t, err)

	t.Run("existing email", func(t *testing.T) {
		user, err := s.Users.GetByEmail(ctx, "test@example.com")
		require.NoError(t, err)
		assert.Equal(t, created.ID, user.ID)
		assert.Equal(t, created.Email, user.Email)
	})

	t.Run("non-existent email", func(t *testing.T) {
		user, err := s.Users.GetByEmail(ctx, "nonexistent@example.com")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, user)
	})
}

func TestUserRepository_GetByUsername(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	created, err := s.Users.Create(ctx, "test@example.com", "testuser", "hash")
	require.NoError(t, err)

	t.Run("existing username", func(t *testing.T) {
		user, err := s.Users.GetByUsername(ctx, "testuser")
		require.NoError(t, err)
		assert.Equal(t, created.ID, user.ID)
		assert.Equal(t, created.Username, user.Username)
	})

	t.Run("non-existent username", func(t *testing.T) {
		user, err := s.Users.GetByUsername(ctx, "nonexistent")
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, user)
	})
}

func TestUserRepository_Update(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	created, err := s.Users.Create(ctx, "test@example.com", "testuser", "hash")
	require.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		updated, err := s.Users.Update(ctx, created.ID, "new@example.com", "newusername")
		require.NoError(t, err)
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, "new@example.com", updated.Email)
		assert.Equal(t, "newusername", updated.Username)
		assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))
	})

	t.Run("update non-existent user", func(t *testing.T) {
		user, err := s.Users.Update(ctx, 99999, "test@example.com", "testuser")
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	created, err := s.Users.Create(ctx, "test@example.com", "testuser", "hash")
	require.NoError(t, err)

	t.Run("successful delete", func(t *testing.T) {
		err = s.Users.Delete(ctx, created.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = s.Users.GetByID(ctx, created.ID)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("delete non-existent user", func(t *testing.T) {
		err = s.Users.Delete(ctx, 99999)
		assert.NoError(t, err) // SQLite doesn't error on DELETE with no rows affected
	})
}

func TestUserRepository_List(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	// Create multiple users
	for i := 0; i < 5; i++ {
		email := fmt.Sprintf("user%d@example.com", i)
		username := fmt.Sprintf("user%d", i)
		_, err := s.Users.Create(ctx, email, username, "hash")
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
			name:      "all users",
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
			users, err := s.Users.List(ctx, tt.limit, tt.offset)
			require.NoError(t, err)
			assert.Len(t, users, tt.wantCount)
		})
	}
}

func TestUserRepository_Delete_CascadeShelves(t *testing.T) {
	t.Parallel()
	s := newTestStorage(t)
	ctx := context.Background()

	// Create user and shelf
	user := createTestUser(t, s, "test@example.com", "testuser")
	shelf := createTestShelf(t, s, user.ID, "My Shelf", false)

	// Delete user
	err := s.Users.Delete(ctx, user.ID)
	require.NoError(t, err)

	// Verify shelf was cascade deleted
	_, err = s.Shelves.GetByID(ctx, shelf.ID)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}
