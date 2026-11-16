package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
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
			repo.EXPECT().TagsFor(mock.Anything, mock.Anything).Return(map[int64][]string{}, nil).Maybe()
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
			repo.EXPECT().TagsFor(mock.Anything, mock.Anything).Return(map[int64][]string{}, nil).Maybe()
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
			repo.EXPECT().TagsFor(mock.Anything, mock.Anything).Return(map[int64][]string{}, nil).Maybe()
			wantLimit := mock.MatchedBy(func(p storage.SearchParams) bool { return p.Limit == tt.want })
			repo.EXPECT().Search(mock.Anything, wantLimit).Return([]db.Book{}, 0, nil)
			s := &Server{
				business: &business.Business{Storage: &storage.Storage{Books: repo, Reviews: noStats(t)}},
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

func TestHandleSearchBooksResponse(t *testing.T) {
	repo := mocks.NewMockBookRepository(t)
	repo.EXPECT().TagsFor(mock.Anything, mock.Anything).Return(map[int64][]string{}, nil).Maybe()
	repo.EXPECT().
		Search(mock.Anything, storage.SearchParams{
			Match: `"hobbit"* author : "tolkien"`, Language: "en", Limit: 5, Offset: 10,
		}).
		Return([]db.Book{{ID: 1, Title: "The Hobbit"}}, int64(11), nil)
	s := &Server{
		business: &business.Business{Storage: &storage.Storage{Books: repo, Reviews: noStats(t)}},
		mux:      http.NewServeMux(),
	}
	s.registerRoutes()

	rec := httptest.NewRecorder()
	url := "/api/books/search?q=hobbit+author:tolkien+lang:en&limit=5&offset=10"
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{`"total":11`, `"limit":5`, `"offset":10`, `"title":"The Hobbit"`} {
		if !strings.Contains(body, want) {
			t.Errorf("body %s lacks %s", body, want)
		}
	}
}

func TestHandleSearchBooksRejectsUnsearchableQuery(t *testing.T) {
	for _, q := range []string{"", "!!!", "%22", "()*", "lang:en"} {
		s := &Server{
			business: &business.Business{Storage: &storage.Storage{Books: mocks.NewMockBookRepository(t)}},
			mux:      http.NewServeMux(),
		}
		s.registerRoutes()
		rec := httptest.NewRecorder()
		s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/books/search?q="+q, nil))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("q=%q: status = %d, want 400", q, rec.Code)
		}
	}
}

// coverRequest sends a create or update request with the given cover_url JSON value to a
// server over a mock repository that accepts both, and returns the response and the
// cover_url the repository received (nil if it was not called).
func coverRequest(t *testing.T, method, cover string) (*httptest.ResponseRecorder, *string) {
	t.Helper()
	var received *string
	repo := mocks.NewMockBookRepository(t)
	repo.EXPECT().TagsFor(mock.Anything, mock.Anything).Return(map[int64][]string{}, nil).Maybe()
	repo.EXPECT().Create(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, p db.CreateBookParams) (*db.Book, error) {
			received = p.CoverUrl
			return &db.Book{ID: 1}, nil
		}).Maybe()
	repo.EXPECT().Update(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, p db.UpdateBookParams) (*db.Book, error) {
			received = p.CoverUrl
			return &db.Book{ID: 1}, nil
		}).Maybe()
	s := &Server{
		business: &business.Business{Storage: &storage.Storage{Books: repo}},
		mux:      http.NewServeMux(),
	}
	s.registerRoutes()

	path := "/api/books"
	if method == http.MethodPut {
		path += "/1"
	}
	body := `{"title":"t","author":"a","cover_url":` + cover + `}`
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec, received
}

func TestHandleBookCoverURL(t *testing.T) {
	long := "https://example.com/" + strings.Repeat("a", 2048-len("https://example.com/"))
	tests := []struct {
		name  string
		cover string // the URL, unquoted
		ok    bool
	}{
		{"http", "http://example.com/a.jpg", true},
		{"https", "https://example.com/a.jpg", true},
		{"exactly 2048", long, true},
		{"too long", long + "a", false},
		{"relative", "/covers/a.jpg", false},
		{"protocol relative", "//example.com/a.jpg", false},
		{"ftp", "ftp://example.com/a.jpg", false},
		{"javascript", "javascript:alert(1)", false},
		{"data", "data:image/png;base64,AAAA", false},
		{"no host", "https://", false},
		{"empty", "", false},
		{"not a url", "not a url", false},
	}
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		for _, tt := range tests {
			t.Run(method+" "+tt.name, func(t *testing.T) {
				rec, received := coverRequest(t, method, `"`+tt.cover+`"`)
				if !tt.ok {
					if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "cover_url") {
						t.Fatalf("got %d %s, want 400 naming cover_url", rec.Code, rec.Body.String())
					}
					return
				}
				if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
					t.Fatalf("status = %d (%s)", rec.Code, rec.Body.String())
				}
				if received == nil || *received != tt.cover {
					t.Errorf("repository received %v, want %q", received, tt.cover)
				}
			})
		}
	}
}

func TestHandleBookCoverURLOmitted(t *testing.T) {
	repo := mocks.NewMockBookRepository(t)
	repo.EXPECT().TagsFor(mock.Anything, mock.Anything).Return(map[int64][]string{}, nil).Maybe()
	repo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(p db.UpdateBookParams) bool {
		return p.CoverUrl == nil
	})).Return(&db.Book{ID: 1}, nil)
	s := &Server{
		business: &business.Business{Storage: &storage.Storage{Books: repo}},
		mux:      http.NewServeMux(),
	}
	s.registerRoutes()

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/api/books/1", strings.NewReader(`{"title":"t","author":"a"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHandleSearchBooksReviewStats(t *testing.T) {
	books := mocks.NewMockBookRepository(t)
	books.EXPECT().TagsFor(mock.Anything, mock.Anything).Return(map[int64][]string{}, nil)
	books.EXPECT().Search(mock.Anything, mock.Anything).
		Return([]db.Book{{ID: 1, Title: "A"}, {ID: 2, Title: "B"}, {ID: 3, Title: "C"}}, int64(3), nil)
	reviews := mocks.NewMockReviewRepository(t)
	// One call for all results, not one per result.
	reviews.EXPECT().Stats(mock.Anything, []int64{1, 2, 3}).Return([]db.GetReviewStatsRow{
		{BookID: 1, AverageRating: 4.25, ReviewCount: 4},
		{BookID: 3, AverageRating: 2, ReviewCount: 1},
	}, nil).Once()
	s := &Server{
		business: &business.Business{Storage: &storage.Storage{Books: books, Reviews: reviews}},
		mux:      http.NewServeMux(),
	}
	s.registerRoutes()

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/books/search?q=a", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	var got struct {
		Books []struct {
			ID            int64    `json:"id"`
			AverageRating *float64 `json:"average_rating"`
			ReviewCount   int64    `json:"review_count"`
		} `json:"books"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	type stat struct {
		Avg   *float64
		Count int64
	}
	want := []stat{{ptr(4.3), 4}, {nil, 0}, {ptr(2), 1}}
	if len(got.Books) != len(want) {
		t.Fatalf("got %d books", len(got.Books))
	}
	for i, w := range want {
		g := got.Books[i]
		if !reflect.DeepEqual(g.AverageRating, w.Avg) || g.ReviewCount != w.Count {
			t.Errorf("book %d = %v/%d, want %v/%d", i+1, g.AverageRating, g.ReviewCount, w.Avg, w.Count)
		}
	}
}

// noStats is a review repository that reports no reviews for any book.
func noStats(t *testing.T) *mocks.MockReviewRepository {
	t.Helper()
	r := mocks.NewMockReviewRepository(t)
	r.EXPECT().Stats(mock.Anything, mock.Anything).Return([]db.GetReviewStatsRow{}, nil).Maybe()
	return r
}
