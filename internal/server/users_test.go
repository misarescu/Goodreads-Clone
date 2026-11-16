package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

func monthlyServer(users *mocks.MockUserRepository, sbs *mocks.MockShelfBookRepository) *Server {
	s := &Server{
		business: &business.Business{Storage: &storage.Storage{Users: users, ShelfBooks: sbs}},
		mux:      http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func TestHandleGetUserMonthlyStats(t *testing.T) {
	year := int64(time.Now().UTC().Year())
	const base = "/api/users/1/stats/monthly"
	tests := []struct {
		name     string
		query    string
		wantYear int64
	}{
		{"explicit year", "?year=2024", 2024},
		{"default year", "", year},
		{"next year", fmt.Sprintf("?year=%d", year+1), year + 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := mocks.NewMockUserRepository(t)
			sbs := mocks.NewMockShelfBookRepository(t)
			users.EXPECT().GetByID(mock.Anything, int64(1)).Return(&db.User{ID: 1}, nil)
			sbs.EXPECT().ListMonthlyFinished(mock.Anything, int64(1), tt.wantYear).
				Return([]db.ListMonthlyFinishedBooksRow{{Month: 3, BooksFinished: 2, PagesFinished: 300}}, nil)

			rec := httptest.NewRecorder()
			monthlyServer(users, sbs).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, base+tt.query, nil))

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			var got []map[string]int64
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if len(got) != 12 {
				t.Fatalf("len = %d, want 12", len(got))
			}
			for i, m := range got {
				if m["month"] != int64(i+1) {
					t.Errorf("entry %d month = %d", i, m["month"])
				}
			}
			if got[2]["books_finished"] != 2 || got[2]["pages_finished"] != 300 || got[0]["books_finished"] != 0 {
				t.Errorf("unexpected entries: %v", got)
			}
		})
	}
}

func TestHandleGetUserMonthlyStats_Errors(t *testing.T) {
	year := time.Now().UTC().Year()
	tests := []struct {
		name       string
		path       string
		userErr    error
		wantStatus int
	}{
		{"non-numeric year", "/api/users/1/stats/monthly?year=abc", nil, http.StatusBadRequest},
		{"too early", "/api/users/1/stats/monthly?year=1899", nil, http.StatusBadRequest},
		{"too late", fmt.Sprintf("/api/users/1/stats/monthly?year=%d", year+2), nil, http.StatusBadRequest},
		{"invalid id", "/api/users/abc/stats/monthly", nil, http.StatusBadRequest},
		{"unknown user", "/api/users/1/stats/monthly", sql.ErrNoRows, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := mocks.NewMockUserRepository(t)
			sbs := mocks.NewMockShelfBookRepository(t)
			if tt.userErr != nil {
				users.EXPECT().GetByID(mock.Anything, int64(1)).Return(nil, tt.userErr)
			}
			rec := httptest.NewRecorder()
			monthlyServer(users, sbs).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
