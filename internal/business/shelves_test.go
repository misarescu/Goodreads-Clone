package business_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goodreads-clone.com/internal/business"
	"goodreads-clone.com/internal/db"
	"goodreads-clone.com/internal/mocks"
	"goodreads-clone.com/internal/storage"
)

func TestBusiness_CreateDefaultShelves(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := int64(1)

	t.Run("successfully creates all default shelves", func(t *testing.T) {
		mockShelfRepo := mocks.NewMockShelfRepository(t)

		// Expect 4 calls to Create, one for each default shelf
		expectedNames := []string{"To Read", "Currently Reading", "Finished", "DNF"}
		for i, name := range expectedNames {
			mockShelfRepo.EXPECT().
				Create(ctx, userID, name, true).
				Return(&db.Shelf{
					ID:        int64(i + 1),
					UserID:    userID,
					Name:      name,
					IsDefault: true,
				}, nil).
				Once()
		}

		b := &business.Business{
			Storage: &storage.Storage{
				Shelves: mockShelfRepo,
			},
		}

		shelves, err := b.CreateDefaultShelves(ctx, userID)

		require.NoError(t, err)
		assert.Len(t, shelves, 4)

		// Verify shelf names match expected order
		for i, name := range expectedNames {
			assert.Equal(t, name, shelves[i].Name)
			assert.True(t, shelves[i].IsDefault)
			assert.Equal(t, userID, shelves[i].UserID)
		}
	})

	t.Run("returns error when first shelf creation fails", func(t *testing.T) {
		mockShelfRepo := mocks.NewMockShelfRepository(t)

		expectedErr := errors.New("database error")
		mockShelfRepo.EXPECT().
			Create(ctx, userID, "To Read", true).
			Return(nil, expectedErr).
			Once()

		b := &business.Business{
			Storage: &storage.Storage{
				Shelves: mockShelfRepo,
			},
		}

		shelves, err := b.CreateDefaultShelves(ctx, userID)

		require.Error(t, err)
		assert.Nil(t, shelves)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("returns error when second shelf creation fails", func(t *testing.T) {
		mockShelfRepo := mocks.NewMockShelfRepository(t)

		// First shelf succeeds
		mockShelfRepo.EXPECT().
			Create(ctx, userID, "To Read", true).
			Return(&db.Shelf{ID: 1, Name: "To Read"}, nil).
			Once()

		// Second shelf fails
		expectedErr := errors.New("database error")
		mockShelfRepo.EXPECT().
			Create(ctx, userID, "Currently Reading", true).
			Return(nil, expectedErr).
			Once()

		b := &business.Business{
			Storage: &storage.Storage{
				Shelves: mockShelfRepo,
			},
		}

		shelves, err := b.CreateDefaultShelves(ctx, userID)

		require.Error(t, err)
		assert.Nil(t, shelves)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("verifies DefaultShelfNames constant", func(t *testing.T) {
		// This test ensures the DefaultShelfNames constant hasn't changed
		assert.Equal(t, []string{"To Read", "Currently Reading", "Finished", "DNF"}, business.DefaultShelfNames)
	})
}

func TestBusiness_CreateShelf(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := int64(1)
	shelfName := "Favorites"

	t.Run("creates custom shelf successfully", func(t *testing.T) {
		mockShelfRepo := mocks.NewMockShelfRepository(t)

		mockShelfRepo.EXPECT().
			Create(ctx, userID, shelfName, false).
			Return(&db.Shelf{
				ID:        5,
				UserID:    userID,
				Name:      shelfName,
				IsDefault: false,
			}, nil).
			Once()

		b := &business.Business{
			Storage: &storage.Storage{
				Shelves: mockShelfRepo,
			},
		}

		shelf, err := b.CreateShelf(ctx, userID, shelfName)

		require.NoError(t, err)
		assert.Equal(t, shelfName, shelf.Name)
		assert.False(t, shelf.IsDefault)
	})

	t.Run("error propagated", func(t *testing.T) {
		mockShelfRepo := mocks.NewMockShelfRepository(t)

		mockShelfRepo.EXPECT().
			Create(ctx, userID, shelfName, false).
			Return(nil, errors.New("duplicate shelf name")).
			Once()

		b := &business.Business{
			Storage: &storage.Storage{
				Shelves: mockShelfRepo,
			},
		}

		shelf, err := b.CreateShelf(ctx, userID, shelfName)

		require.Error(t, err)
		assert.Nil(t, shelf)
	})
}

func TestBusiness_GetShelfByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	shelfID := int64(1)

	tests := []struct {
		name        string
		mockSetup   func(*mocks.MockShelfRepository)
		wantErr     bool
		checkResult func(*testing.T, *db.Shelf)
	}{
		{
			name: "shelf exists",
			mockSetup: func(m *mocks.MockShelfRepository) {
				m.EXPECT().
					GetByID(ctx, shelfID).
					Return(&db.Shelf{ID: shelfID, Name: "To Read"}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, shelf *db.Shelf) {
				assert.Equal(t, shelfID, shelf.ID)
				assert.Equal(t, "To Read", shelf.Name)
			},
		},
		{
			name: "shelf not found",
			mockSetup: func(m *mocks.MockShelfRepository) {
				m.EXPECT().
					GetByID(ctx, shelfID).
					Return(nil, sql.ErrNoRows).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShelfRepo := mocks.NewMockShelfRepository(t)
			tt.mockSetup(mockShelfRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					Shelves: mockShelfRepo,
				},
			}

			shelf, err := b.GetShelfByID(ctx, shelfID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, shelf)
			}
		})
	}
}

func TestBusiness_GetShelfByUserAndName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := int64(1)
	shelfName := "Favorites"

	tests := []struct {
		name        string
		mockSetup   func(*mocks.MockShelfRepository)
		wantErr     bool
		checkResult func(*testing.T, *db.Shelf)
	}{
		{
			name: "shelf found",
			mockSetup: func(m *mocks.MockShelfRepository) {
				m.EXPECT().
					GetByUserAndName(ctx, userID, shelfName).
					Return(&db.Shelf{ID: 1, UserID: userID, Name: shelfName}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, shelf *db.Shelf) {
				assert.Equal(t, shelfName, shelf.Name)
			},
		},
		{
			name: "shelf not found",
			mockSetup: func(m *mocks.MockShelfRepository) {
				m.EXPECT().
					GetByUserAndName(ctx, userID, shelfName).
					Return(nil, sql.ErrNoRows).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShelfRepo := mocks.NewMockShelfRepository(t)
			tt.mockSetup(mockShelfRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					Shelves: mockShelfRepo,
				},
			}

			shelf, err := b.GetShelfByUserAndName(ctx, userID, shelfName)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, shelf)
			}
		})
	}
}

func TestBusiness_ListShelvesByUser(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := int64(1)

	expectedShelves := []db.Shelf{
		{ID: 1, UserID: userID, Name: "To Read", IsDefault: true},
		{ID: 2, UserID: userID, Name: "Favorites", IsDefault: false},
	}

	mockShelfRepo := mocks.NewMockShelfRepository(t)
	mockShelfRepo.EXPECT().
		ListByUser(ctx, userID).
		Return(expectedShelves, nil).
		Once()

	b := &business.Business{
		Storage: &storage.Storage{
			Shelves: mockShelfRepo,
		},
	}

	shelves, err := b.ListShelvesByUser(ctx, userID)

	require.NoError(t, err)
	assert.Len(t, shelves, 2)
	assert.Equal(t, expectedShelves, shelves)
}

func TestBusiness_UpdateShelf(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	shelfID := int64(1)
	newName := "Updated Name"

	tests := []struct {
		name        string
		mockSetup   func(*mocks.MockShelfRepository)
		wantErr     bool
		checkResult func(*testing.T, *db.Shelf)
	}{
		{
			name: "successful update",
			mockSetup: func(m *mocks.MockShelfRepository) {
				m.EXPECT().
					Update(ctx, shelfID, newName).
					Return(&db.Shelf{ID: shelfID, Name: newName}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, shelf *db.Shelf) {
				assert.Equal(t, newName, shelf.Name)
			},
		},
		{
			name: "shelf not found",
			mockSetup: func(m *mocks.MockShelfRepository) {
				m.EXPECT().
					Update(ctx, shelfID, newName).
					Return(nil, sql.ErrNoRows).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShelfRepo := mocks.NewMockShelfRepository(t)
			tt.mockSetup(mockShelfRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					Shelves: mockShelfRepo,
				},
			}

			shelf, err := b.UpdateShelf(ctx, shelfID, newName)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, shelf)
			}
		})
	}
}

func TestBusiness_DeleteShelf(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	shelfID := int64(1)

	tests := []struct {
		name      string
		mockSetup func(*mocks.MockShelfRepository)
		wantErr   bool
	}{
		{
			name: "successful delete",
			mockSetup: func(m *mocks.MockShelfRepository) {
				m.EXPECT().
					Delete(ctx, shelfID).
					Return(nil).
					Once()
			},
			wantErr: false,
		},
		{
			name: "error propagated",
			mockSetup: func(m *mocks.MockShelfRepository) {
				m.EXPECT().
					Delete(ctx, shelfID).
					Return(errors.New("database error")).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShelfRepo := mocks.NewMockShelfRepository(t)
			tt.mockSetup(mockShelfRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					Shelves: mockShelfRepo,
				},
			}

			err := b.DeleteShelf(ctx, shelfID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBusiness_GetDefaultShelvesByUser(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := int64(1)

	expectedShelves := []db.Shelf{
		{ID: 1, UserID: userID, Name: "To Read", IsDefault: true},
		{ID: 2, UserID: userID, Name: "Currently Reading", IsDefault: true},
	}

	mockShelfRepo := mocks.NewMockShelfRepository(t)
	mockShelfRepo.EXPECT().
		GetDefaultsByUser(ctx, userID).
		Return(expectedShelves, nil).
		Once()

	b := &business.Business{
		Storage: &storage.Storage{
			Shelves: mockShelfRepo,
		},
	}

	shelves, err := b.GetDefaultShelvesByUser(ctx, userID)

	require.NoError(t, err)
	assert.Len(t, shelves, 2)
	for _, shelf := range shelves {
		assert.True(t, shelf.IsDefault)
	}
}
