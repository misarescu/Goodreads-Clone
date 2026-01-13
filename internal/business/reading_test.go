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

func TestBusiness_AddBookToShelf(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	shelfID := int64(1)
	bookID := int64(2)

	tests := []struct {
		name            string
		pagesRead       int64
		progressPercent int64
		mockSetup       func(*mocks.MockShelfBookRepository)
		wantErr         bool
		checkResult     func(*testing.T, *db.ShelfBook)
	}{
		{
			name:            "successful add without progress",
			pagesRead:       0,
			progressPercent: 0,
			mockSetup: func(m *mocks.MockShelfBookRepository) {
				m.EXPECT().
					Add(ctx, shelfID, bookID, int64(0), int64(0)).
					Return(&db.ShelfBook{
						ID:              1,
						ShelfID:         shelfID,
						BookID:          bookID,
						PagesRead:       0,
						ProgressPercent: 0,
					}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, sb *db.ShelfBook) {
				assert.Equal(t, shelfID, sb.ShelfID)
				assert.Equal(t, bookID, sb.BookID)
				assert.Equal(t, int64(0), sb.PagesRead)
				assert.Equal(t, int64(0), sb.ProgressPercent)
			},
		},
		{
			name:            "successful add with progress",
			pagesRead:       150,
			progressPercent: 50,
			mockSetup: func(m *mocks.MockShelfBookRepository) {
				m.EXPECT().
					Add(ctx, shelfID, bookID, int64(150), int64(50)).
					Return(&db.ShelfBook{
						ID:              1,
						ShelfID:         shelfID,
						BookID:          bookID,
						PagesRead:       150,
						ProgressPercent: 50,
					}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, sb *db.ShelfBook) {
				assert.Equal(t, int64(150), sb.PagesRead)
				assert.Equal(t, int64(50), sb.ProgressPercent)
			},
		},
		{
			name:            "duplicate shelf_book error",
			pagesRead:       0,
			progressPercent: 0,
			mockSetup: func(m *mocks.MockShelfBookRepository) {
				m.EXPECT().
					Add(ctx, shelfID, bookID, mock.Anything, mock.Anything).
					Return(nil, errors.New("UNIQUE constraint failed")).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShelfBookRepo := mocks.NewMockShelfBookRepository(t)
			tt.mockSetup(mockShelfBookRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					ShelfBooks: mockShelfBookRepo,
				},
			}

			shelfBook, err := b.AddBookToShelf(ctx, shelfID, bookID, tt.pagesRead, tt.progressPercent)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, shelfBook)
			}
		})
	}
}

func TestBusiness_GetShelfBook(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	shelfID := int64(1)
	bookID := int64(2)

	tests := []struct {
		name        string
		mockSetup   func(*mocks.MockShelfBookRepository)
		wantErr     bool
		checkResult func(*testing.T, *db.ShelfBook)
	}{
		{
			name: "shelf_book found",
			mockSetup: func(m *mocks.MockShelfBookRepository) {
				m.EXPECT().
					Get(ctx, shelfID, bookID).
					Return(&db.ShelfBook{
						ID:              1,
						ShelfID:         shelfID,
						BookID:          bookID,
						PagesRead:       100,
						ProgressPercent: 25,
					}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, sb *db.ShelfBook) {
				assert.Equal(t, shelfID, sb.ShelfID)
				assert.Equal(t, bookID, sb.BookID)
				assert.Equal(t, int64(100), sb.PagesRead)
			},
		},
		{
			name: "shelf_book not found",
			mockSetup: func(m *mocks.MockShelfBookRepository) {
				m.EXPECT().
					Get(ctx, shelfID, bookID).
					Return(nil, sql.ErrNoRows).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShelfBookRepo := mocks.NewMockShelfBookRepository(t)
			tt.mockSetup(mockShelfBookRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					ShelfBooks: mockShelfBookRepo,
				},
			}

			shelfBook, err := b.GetShelfBook(ctx, shelfID, bookID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, shelfBook)
			}
		})
	}
}

func TestBusiness_UpdateReadingProgress(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	shelfBookID := int64(1)

	tests := []struct {
		name            string
		pagesRead       int64
		progressPercent int64
		mockSetup       func(*mocks.MockShelfBookRepository)
		wantErr         bool
		checkResult     func(*testing.T, *db.ShelfBook)
	}{
		{
			name:            "successful update",
			pagesRead:       200,
			progressPercent: 50,
			mockSetup: func(m *mocks.MockShelfBookRepository) {
				m.EXPECT().
					UpdateProgress(ctx, shelfBookID, int64(200), int64(50)).
					Return(&db.ShelfBook{
						ID:              shelfBookID,
						PagesRead:       200,
						ProgressPercent: 50,
					}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, sb *db.ShelfBook) {
				assert.Equal(t, int64(200), sb.PagesRead)
				assert.Equal(t, int64(50), sb.ProgressPercent)
			},
		},
		{
			name:            "update to completion",
			pagesRead:       400,
			progressPercent: 100,
			mockSetup: func(m *mocks.MockShelfBookRepository) {
				m.EXPECT().
					UpdateProgress(ctx, shelfBookID, int64(400), int64(100)).
					Return(&db.ShelfBook{
						ID:              shelfBookID,
						PagesRead:       400,
						ProgressPercent: 100,
					}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, sb *db.ShelfBook) {
				assert.Equal(t, int64(400), sb.PagesRead)
				assert.Equal(t, int64(100), sb.ProgressPercent)
			},
		},
		{
			name:            "shelf_book not found",
			pagesRead:       100,
			progressPercent: 50,
			mockSetup: func(m *mocks.MockShelfBookRepository) {
				m.EXPECT().
					UpdateProgress(ctx, shelfBookID, mock.Anything, mock.Anything).
					Return(nil, sql.ErrNoRows).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShelfBookRepo := mocks.NewMockShelfBookRepository(t)
			tt.mockSetup(mockShelfBookRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					ShelfBooks: mockShelfBookRepo,
				},
			}

			shelfBook, err := b.UpdateReadingProgress(ctx, shelfBookID, tt.pagesRead, tt.progressPercent)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, shelfBook)
			}
		})
	}
}

func TestBusiness_RemoveBookFromShelf(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	shelfID := int64(1)
	bookID := int64(2)

	tests := []struct {
		name      string
		mockSetup func(*mocks.MockShelfBookRepository)
		wantErr   bool
	}{
		{
			name: "successful remove",
			mockSetup: func(m *mocks.MockShelfBookRepository) {
				m.EXPECT().
					Remove(ctx, shelfID, bookID).
					Return(nil).
					Once()
			},
			wantErr: false,
		},
		{
			name: "error propagated",
			mockSetup: func(m *mocks.MockShelfBookRepository) {
				m.EXPECT().
					Remove(ctx, shelfID, bookID).
					Return(errors.New("database error")).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShelfBookRepo := mocks.NewMockShelfBookRepository(t)
			tt.mockSetup(mockShelfBookRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					ShelfBooks: mockShelfBookRepo,
				},
			}

			err := b.RemoveBookFromShelf(ctx, shelfID, bookID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBusiness_ListBooksOnShelf(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	shelfID := int64(1)

	tests := []struct {
		name        string
		mockSetup   func(*mocks.MockShelfBookRepository)
		wantErr     bool
		checkResult func(*testing.T, []db.ListBooksOnShelfRow)
	}{
		{
			name: "books found",
			mockSetup: func(m *mocks.MockShelfBookRepository) {
				m.EXPECT().
					ListBooksOnShelf(ctx, shelfID).
					Return([]db.ListBooksOnShelfRow{
						{BookID: 1, Title: "Book 1", Author: "Author 1", ShelfID: shelfID},
						{BookID: 2, Title: "Book 2", Author: "Author 2", ShelfID: shelfID},
					}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, books []db.ListBooksOnShelfRow) {
				assert.Len(t, books, 2)
				assert.Equal(t, "Book 1", books[0].Title)
				assert.Equal(t, "Book 2", books[1].Title)
			},
		},
		{
			name: "empty shelf",
			mockSetup: func(m *mocks.MockShelfBookRepository) {
				m.EXPECT().
					ListBooksOnShelf(ctx, shelfID).
					Return([]db.ListBooksOnShelfRow{}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, books []db.ListBooksOnShelfRow) {
				assert.Len(t, books, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShelfBookRepo := mocks.NewMockShelfBookRepository(t)
			tt.mockSetup(mockShelfBookRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					ShelfBooks: mockShelfBookRepo,
				},
			}

			books, err := b.ListBooksOnShelf(ctx, shelfID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, books)
			}
		})
	}
}

func TestBusiness_GetUserBookShelves(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := int64(1)
	bookID := int64(2)

	tests := []struct {
		name        string
		mockSetup   func(*mocks.MockShelfBookRepository)
		wantErr     bool
		checkResult func(*testing.T, []db.Shelf)
	}{
		{
			name: "book on multiple shelves",
			mockSetup: func(m *mocks.MockShelfBookRepository) {
				m.EXPECT().
					GetUserBookShelves(ctx, userID, bookID).
					Return([]db.Shelf{
						{ID: 1, UserID: userID, Name: "To Read"},
						{ID: 2, UserID: userID, Name: "Favorites"},
					}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, shelves []db.Shelf) {
				assert.Len(t, shelves, 2)
				assert.Equal(t, "To Read", shelves[0].Name)
				assert.Equal(t, "Favorites", shelves[1].Name)
			},
		},
		{
			name: "book not on any shelves",
			mockSetup: func(m *mocks.MockShelfBookRepository) {
				m.EXPECT().
					GetUserBookShelves(ctx, userID, bookID).
					Return([]db.Shelf{}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, shelves []db.Shelf) {
				assert.Len(t, shelves, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShelfBookRepo := mocks.NewMockShelfBookRepository(t)
			tt.mockSetup(mockShelfBookRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					ShelfBooks: mockShelfBookRepo,
				},
			}

			shelves, err := b.GetUserBookShelves(ctx, userID, bookID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, shelves)
			}
		})
	}
}
