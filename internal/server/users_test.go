package server

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"goodreads-clone.com/internal/business"
	"goodreads-clone.com/internal/db"
	"goodreads-clone.com/internal/mocks"
	"goodreads-clone.com/internal/storage"
)

func TestHandleGetUserStats(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		userErr    error
		stats      db.GetUserReadingStatsRow
		wantStatus int
		wantBody   string
	}{
		{
			name:       "user with books",
			path:       "/api/users/1/stats",
			stats:      db.GetUserReadingStatsRow{BooksOnShelves: 3, PagesRead: 450, BooksFinished: 1},
			wantStatus: http.StatusOK,
			wantBody:   `{"books_on_shelves":3,"pages_read":450,"books_finished":1}`,
		},
		{
			name:       "user without shelves",
			path:       "/api/users/1/stats",
			wantStatus: http.StatusOK,
			wantBody:   `{"books_on_shelves":0,"pages_read":0,"books_finished":0}`,
		},
		{
			name:       "unknown user",
			path:       "/api/users/1/stats",
			userErr:    sql.ErrNoRows,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid id",
			path:       "/api/users/abc/stats",
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := mocks.NewMockUserRepository(t)
			sbs := mocks.NewMockShelfBookRepository(t)
			if tt.wantStatus != http.StatusBadRequest {
				if tt.userErr != nil {
					users.EXPECT().GetByID(mock.Anything, int64(1)).Return(nil, tt.userErr)
				} else {
					users.EXPECT().GetByID(mock.Anything, int64(1)).Return(&db.User{ID: 1}, nil)
					sbs.EXPECT().GetUserReadingStats(mock.Anything, int64(1)).Return(&tt.stats, nil)
				}
			}
			s := &Server{
				business: &business.Business{Storage: &storage.Storage{Users: users, ShelfBooks: sbs}},
				mux:      http.NewServeMux(),
			}
			s.registerRoutes()

			rec := httptest.NewRecorder()
			s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" {
				if got := strings.TrimSpace(rec.Body.String()); got != tt.wantBody {
					t.Errorf("body = %q, want %q", got, tt.wantBody)
				}
			}
		})
	}
}
