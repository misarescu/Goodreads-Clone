package business_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"goodreads-clone.com/internal/business"
	"goodreads-clone.com/internal/db"
	"goodreads-clone.com/internal/mocks"
	"goodreads-clone.com/internal/storage"
)

func TestBusiness_CreateUser(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name         string
		email        string
		username     string
		passwordHash string
		mockSetup    func(*mocks.MockUserRepository)
		wantErr      bool
		checkResult  func(*testing.T, *db.User)
	}{
		{
			name:         "successful user creation",
			email:        "test@example.com",
			username:     "testuser",
			passwordHash: "hashed_password",
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().
					Create(ctx, "test@example.com", "testuser", "hashed_password").
					Return(&db.User{
						ID:           1,
						Email:        "test@example.com",
						Username:     "testuser",
						PasswordHash: "hashed_password",
					}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, user *db.User) {
				assert.Equal(t, int64(1), user.ID)
				assert.Equal(t, "test@example.com", user.Email)
				assert.Equal(t, "testuser", user.Username)
			},
		},
		{
			name:         "duplicate email error",
			email:        "test@example.com",
			username:     "testuser",
			passwordHash: "hashed_password",
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().
					Create(ctx, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("UNIQUE constraint failed: users.email")).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewMockUserRepository(t)
			tt.mockSetup(mockUserRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					Users: mockUserRepo,
				},
			}

			user, err := b.CreateUser(ctx, tt.email, tt.username, tt.passwordHash)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, user)
			}
		})
	}
}

func TestBusiness_GetUserByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name        string
		userID      int64
		mockSetup   func(*mocks.MockUserRepository)
		wantErr     bool
		checkResult func(*testing.T, *db.User)
	}{
		{
			name:   "user found",
			userID: 1,
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().
					GetByID(ctx, int64(1)).
					Return(&db.User{
						ID:       1,
						Email:    "test@example.com",
						Username: "testuser",
					}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, user *db.User) {
				assert.Equal(t, int64(1), user.ID)
				assert.Equal(t, "test@example.com", user.Email)
			},
		},
		{
			name:   "user not found",
			userID: 99999,
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().
					GetByID(ctx, int64(99999)).
					Return(nil, sql.ErrNoRows).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewMockUserRepository(t)
			tt.mockSetup(mockUserRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					Users: mockUserRepo,
				},
			}

			user, err := b.GetUserByID(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, user)
			}
		})
	}
}

func TestBusiness_GetUserByEmail(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name        string
		email       string
		mockSetup   func(*mocks.MockUserRepository)
		wantErr     bool
		checkResult func(*testing.T, *db.User)
	}{
		{
			name:  "user found",
			email: "test@example.com",
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().
					GetByEmail(ctx, "test@example.com").
					Return(&db.User{
						ID:       1,
						Email:    "test@example.com",
						Username: "testuser",
					}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, user *db.User) {
				assert.Equal(t, "test@example.com", user.Email)
			},
		},
		{
			name:  "user not found",
			email: "nonexistent@example.com",
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().
					GetByEmail(ctx, "nonexistent@example.com").
					Return(nil, sql.ErrNoRows).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewMockUserRepository(t)
			tt.mockSetup(mockUserRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					Users: mockUserRepo,
				},
			}

			user, err := b.GetUserByEmail(ctx, tt.email)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, user)
			}
		})
	}
}

func TestBusiness_GetUserByUsername(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name        string
		username    string
		mockSetup   func(*mocks.MockUserRepository)
		wantErr     bool
		checkResult func(*testing.T, *db.User)
	}{
		{
			name:     "user found",
			username: "testuser",
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().
					GetByUsername(ctx, "testuser").
					Return(&db.User{
						ID:       1,
						Email:    "test@example.com",
						Username: "testuser",
					}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, user *db.User) {
				assert.Equal(t, "testuser", user.Username)
			},
		},
		{
			name:     "user not found",
			username: "nonexistent",
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().
					GetByUsername(ctx, "nonexistent").
					Return(nil, sql.ErrNoRows).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewMockUserRepository(t)
			tt.mockSetup(mockUserRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					Users: mockUserRepo,
				},
			}

			user, err := b.GetUserByUsername(ctx, tt.username)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, user)
			}
		})
	}
}

func TestBusiness_UpdateUser(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name        string
		userID      int64
		email       string
		username    string
		mockSetup   func(*mocks.MockUserRepository)
		wantErr     bool
		checkResult func(*testing.T, *db.User)
	}{
		{
			name:     "successful update",
			userID:   1,
			email:    "new@example.com",
			username: "newusername",
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().
					Update(ctx, int64(1), "new@example.com", "newusername").
					Return(&db.User{
						ID:       1,
						Email:    "new@example.com",
						Username: "newusername",
					}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, user *db.User) {
				assert.Equal(t, "new@example.com", user.Email)
				assert.Equal(t, "newusername", user.Username)
			},
		},
		{
			name:     "user not found",
			userID:   99999,
			email:    "test@example.com",
			username: "testuser",
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().
					Update(ctx, int64(99999), mock.Anything, mock.Anything).
					Return(nil, sql.ErrNoRows).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewMockUserRepository(t)
			tt.mockSetup(mockUserRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					Users: mockUserRepo,
				},
			}

			user, err := b.UpdateUser(ctx, tt.userID, tt.email, tt.username)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, user)
			}
		})
	}
}

func TestBusiness_DeleteUser(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name      string
		userID    int64
		mockSetup func(*mocks.MockUserRepository)
		wantErr   bool
	}{
		{
			name:   "successful delete",
			userID: 1,
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().
					Delete(ctx, int64(1)).
					Return(nil).
					Once()
			},
			wantErr: false,
		},
		{
			name:   "delete error propagated",
			userID: 1,
			mockSetup: func(m *mocks.MockUserRepository) {
				m.EXPECT().
					Delete(ctx, int64(1)).
					Return(errors.New("database error")).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewMockUserRepository(t)
			tt.mockSetup(mockUserRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					Users: mockUserRepo,
				},
			}

			err := b.DeleteUser(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBusiness_ListUsers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	expectedUsers := []db.User{
		{ID: 1, Email: "user1@example.com", Username: "user1"},
		{ID: 2, Email: "user2@example.com", Username: "user2"},
	}

	mockUserRepo := mocks.NewMockUserRepository(t)
	mockUserRepo.EXPECT().
		List(ctx, int64(10), int64(0)).
		Return(expectedUsers, nil).
		Once()

	b := &business.Business{
		Storage: &storage.Storage{
			Users: mockUserRepo,
		},
	}

	users, err := b.ListUsers(ctx, 10, 0)

	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, expectedUsers, users)
}
