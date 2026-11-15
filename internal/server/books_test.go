package server

import (
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

func TestHandleCountBooks(t *testing.T) {
	for name, n := range map[string]int64{"empty catalog": 0, "catalog with books": 3} {
		t.Run(name, func(t *testing.T) {
			repo := mocks.NewMockBookRepository(t)
			repo.EXPECT().Count(mock.Anything).Return(n, nil)
			s := &Server{
				business: &business.Business{Storage: &storage.Storage{Books: repo}},
				mux:      http.NewServeMux(),
			}
			s.registerRoutes()

			rec := httptest.NewRecorder()
			s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/books/count", nil))

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); ct != contentTypeJSON {
				t.Errorf("Content-Type = %q", ct)
			}
			if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", cc)
			}
			want := map[int64]string{0: `{"book_count":0}`, 3: `{"book_count":3}`}[n]
			if got := strings.TrimSpace(rec.Body.String()); got != want {
				t.Errorf("body = %q, want %q", got, want)
			}
		})
	}
}

func TestHandleUpdateBookRating(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantRating *float64
	}{
		{"quarter step", `{"title":"t","author":"a","rating":3.75}`, http.StatusOK, ptr(3.75)},
		{"zero", `{"title":"t","author":"a","rating":0}`, http.StatusOK, ptr(0.0)},
		{"max", `{"title":"t","author":"a","rating":5}`, http.StatusOK, ptr(5.0)},
		{"omitted", `{"title":"t","author":"a"}`, http.StatusOK, nil},
		{"off increment", `{"title":"t","author":"a","rating":3.3}`, http.StatusBadRequest, nil},
		{"too high", `{"title":"t","author":"a","rating":5.25}`, http.StatusBadRequest, nil},
		{"negative", `{"title":"t","author":"a","rating":-0.25}`, http.StatusBadRequest, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockBookRepository(t)
			if tt.wantStatus == http.StatusOK {
				repo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(p db.UpdateBookParams) bool {
					if tt.wantRating == nil {
						return !p.Rating.Valid
					}
					return p.Rating.Valid && p.Rating.Float64 == *tt.wantRating
				})).Return(&db.Book{ID: 1}, nil)
			}
			s := &Server{
				business: &business.Business{Storage: &storage.Storage{Books: repo}},
				mux:      http.NewServeMux(),
			}
			s.registerRoutes()

			rec := httptest.NewRecorder()
			s.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/books/1", strings.NewReader(tt.body)))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func ptr(f float64) *float64 { return &f }

func TestHandleSearchBooksLimit(t *testing.T) {
	tests := map[string]struct {
		query string
		want  int64
	}{
		"default":       {"/api/books/search?q=a", 20},
		"within cap":    {"/api/books/search?q=a&limit=50", 50},
		"above cap":     {"/api/books/search?q=a&limit=1000000", 100},
		"invalid limit": {"/api/books/search?q=a&limit=-5", 20},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repo := mocks.NewMockBookRepository(t)
			repo.EXPECT().SearchByTitle(mock.Anything, "a", tt.want).Return([]db.Book{}, nil)
			s := &Server{
				business: &business.Business{Storage: &storage.Storage{Books: repo}},
				mux:      http.NewServeMux(),
			}
			s.registerRoutes()

			rec := httptest.NewRecorder()
			s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.query, nil))
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
		})
	}
}
