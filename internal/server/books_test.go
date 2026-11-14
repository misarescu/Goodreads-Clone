package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"goodreads-clone.com/internal/business"
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
