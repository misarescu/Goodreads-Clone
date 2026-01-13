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

func TestBusiness_CreateBook(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	params := db.CreateBookParams{
		Title:  "Test Book",
		Author: "Test Author",
	}

	tests := []struct {
		name        string
		params      db.CreateBookParams
		mockSetup   func(*mocks.MockBookRepository)
		wantErr     bool
		checkResult func(*testing.T, *db.Book)
	}{
		{
			name:   "successful book creation",
			params: params,
			mockSetup: func(m *mocks.MockBookRepository) {
				m.EXPECT().
					Create(ctx, params).
					Return(&db.Book{
						ID:     1,
						Title:  "Test Book",
						Author: "Test Author",
					}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, book *db.Book) {
				assert.Equal(t, int64(1), book.ID)
				assert.Equal(t, "Test Book", book.Title)
				assert.Equal(t, "Test Author", book.Author)
			},
		},
		{
			name:   "duplicate ISBN error",
			params: params,
			mockSetup: func(m *mocks.MockBookRepository) {
				m.EXPECT().
					Create(ctx, mock.Anything).
					Return(nil, errors.New("UNIQUE constraint failed: books.isbn")).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBookRepo := mocks.NewMockBookRepository(t)
			tt.mockSetup(mockBookRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					Books: mockBookRepo,
				},
			}

			book, err := b.CreateBook(ctx, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, book)
			}
		})
	}
}

func TestBusiness_GetBookByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name        string
		bookID      int64
		mockSetup   func(*mocks.MockBookRepository)
		wantErr     bool
		checkResult func(*testing.T, *db.Book)
	}{
		{
			name:   "book found",
			bookID: 1,
			mockSetup: func(m *mocks.MockBookRepository) {
				m.EXPECT().
					GetByID(ctx, int64(1)).
					Return(&db.Book{
						ID:     1,
						Title:  "Found Book",
						Author: "Author",
					}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, book *db.Book) {
				assert.Equal(t, int64(1), book.ID)
				assert.Equal(t, "Found Book", book.Title)
			},
		},
		{
			name:   "book not found",
			bookID: 99999,
			mockSetup: func(m *mocks.MockBookRepository) {
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
			mockBookRepo := mocks.NewMockBookRepository(t)
			tt.mockSetup(mockBookRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					Books: mockBookRepo,
				},
			}

			book, err := b.GetBookByID(ctx, tt.bookID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, book)
			}
		})
	}
}

func TestBusiness_GetBookByISBN(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name        string
		isbn        string
		mockSetup   func(*mocks.MockBookRepository)
		wantErr     bool
		checkResult func(*testing.T, *db.Book)
	}{
		{
			name: "book found",
			isbn: "978-1234567890",
			mockSetup: func(m *mocks.MockBookRepository) {
				m.EXPECT().
					GetByISBN(ctx, "978-1234567890").
					Return(&db.Book{
						ID:     1,
						Title:  "ISBN Book",
						Author: "Author",
						Isbn: sql.NullString{
							String: "978-1234567890",
							Valid:  true,
						},
					}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, book *db.Book) {
				assert.Equal(t, "978-1234567890", book.Isbn.String)
			},
		},
		{
			name: "book not found",
			isbn: "978-0000000000",
			mockSetup: func(m *mocks.MockBookRepository) {
				m.EXPECT().
					GetByISBN(ctx, "978-0000000000").
					Return(nil, sql.ErrNoRows).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBookRepo := mocks.NewMockBookRepository(t)
			tt.mockSetup(mockBookRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					Books: mockBookRepo,
				},
			}

			book, err := b.GetBookByISBN(ctx, tt.isbn)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, book)
			}
		})
	}
}

func TestBusiness_UpdateBook(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	params := db.UpdateBookParams{
		ID:     1,
		Title:  "Updated Title",
		Author: "Updated Author",
	}

	tests := []struct {
		name        string
		params      db.UpdateBookParams
		mockSetup   func(*mocks.MockBookRepository)
		wantErr     bool
		checkResult func(*testing.T, *db.Book)
	}{
		{
			name:   "successful update",
			params: params,
			mockSetup: func(m *mocks.MockBookRepository) {
				m.EXPECT().
					Update(ctx, params).
					Return(&db.Book{
						ID:     1,
						Title:  "Updated Title",
						Author: "Updated Author",
					}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, book *db.Book) {
				assert.Equal(t, "Updated Title", book.Title)
				assert.Equal(t, "Updated Author", book.Author)
			},
		},
		{
			name:   "book not found",
			params: params,
			mockSetup: func(m *mocks.MockBookRepository) {
				m.EXPECT().
					Update(ctx, mock.Anything).
					Return(nil, sql.ErrNoRows).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBookRepo := mocks.NewMockBookRepository(t)
			tt.mockSetup(mockBookRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					Books: mockBookRepo,
				},
			}

			book, err := b.UpdateBook(ctx, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, book)
			}
		})
	}
}

func TestBusiness_DeleteBook(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name      string
		bookID    int64
		mockSetup func(*mocks.MockBookRepository)
		wantErr   bool
	}{
		{
			name:   "successful delete",
			bookID: 1,
			mockSetup: func(m *mocks.MockBookRepository) {
				m.EXPECT().
					Delete(ctx, int64(1)).
					Return(nil).
					Once()
			},
			wantErr: false,
		},
		{
			name:   "error propagated",
			bookID: 1,
			mockSetup: func(m *mocks.MockBookRepository) {
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
			mockBookRepo := mocks.NewMockBookRepository(t)
			tt.mockSetup(mockBookRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					Books: mockBookRepo,
				},
			}

			err := b.DeleteBook(ctx, tt.bookID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBusiness_ListBooks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	expectedBooks := []db.Book{
		{ID: 1, Title: "Book 1", Author: "Author 1"},
		{ID: 2, Title: "Book 2", Author: "Author 2"},
	}

	mockBookRepo := mocks.NewMockBookRepository(t)
	mockBookRepo.EXPECT().
		List(ctx, int64(10), int64(0)).
		Return(expectedBooks, nil).
		Once()

	b := &business.Business{
		Storage: &storage.Storage{
			Books: mockBookRepo,
		},
	}

	books, err := b.ListBooks(ctx, 10, 0)

	require.NoError(t, err)
	assert.Len(t, books, 2)
	assert.Equal(t, expectedBooks, books)
}

func TestBusiness_SearchBooksByTitle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name        string
		query       string
		limit       int64
		mockSetup   func(*mocks.MockBookRepository)
		wantErr     bool
		checkResult func(*testing.T, []db.Book)
	}{
		{
			name:  "books found",
			query: "Great",
			limit: 10,
			mockSetup: func(m *mocks.MockBookRepository) {
				m.EXPECT().
					SearchByTitle(ctx, "Great", int64(10)).
					Return([]db.Book{
						{ID: 1, Title: "The Great Gatsby", Author: "F. Scott Fitzgerald"},
						{ID: 2, Title: "Great Expectations", Author: "Charles Dickens"},
					}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, books []db.Book) {
				assert.Len(t, books, 2)
			},
		},
		{
			name:  "no results",
			query: "Nonexistent",
			limit: 10,
			mockSetup: func(m *mocks.MockBookRepository) {
				m.EXPECT().
					SearchByTitle(ctx, "Nonexistent", int64(10)).
					Return([]db.Book{}, nil).
					Once()
			},
			wantErr: false,
			checkResult: func(t *testing.T, books []db.Book) {
				assert.Empty(t, books)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBookRepo := mocks.NewMockBookRepository(t)
			tt.mockSetup(mockBookRepo)

			b := &business.Business{
				Storage: &storage.Storage{
					Books: mockBookRepo,
				},
			}

			books, err := b.SearchBooksByTitle(ctx, tt.query, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.checkResult(t, books)
			}
		})
	}
}
