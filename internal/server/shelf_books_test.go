package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"goodreads-clone.com/internal/business"
	"goodreads-clone.com/internal/db"
	"goodreads-clone.com/internal/storage"

	_ "modernc.org/sqlite"
)

// newShelfBooksServer builds a server on a real in-memory SQLite database so
// that constraint errors (unique, foreign key) are the genuine driver errors.
// It seeds user 1, shelf 1 and book 1.
func newShelfBooksServer(t *testing.T) *Server {
	t.Helper()

	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	database.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = database.Close() })

	if _, err := database.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join("..", "..", "db", "migrations", "*.sql"))
	if err != nil || len(files) == 0 {
		t.Fatalf("migrations not found: %v", err)
	}
	sort.Strings(files)
	for _, f := range files {
		sqlText, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := database.Exec(string(sqlText)); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
	}
	for _, q := range []string{
		`INSERT INTO users (email, username, password_hash) VALUES ('a@b.c', 'u', 'x')`,
		`INSERT INTO shelves (user_id, name) VALUES (1, 'Shelf')`,
		`INSERT INTO books (title, author) VALUES ('Title', 'Author')`,
	} {
		if _, err := database.Exec(q); err != nil {
			t.Fatalf("seed %q: %v", q, err)
		}
	}

	s := &Server{
		business: &business.Business{Storage: storage.New(database)},
		mux:      http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func doRequest(s *Server, method, path, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
	}
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	return rec
}

func assertNewShelfBook(t *testing.T, body []byte) {
	t.Helper()
	var got db.ShelfBook
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.ShelfID != 1 || got.BookID != 1 || got.PagesRead != 0 || got.ProgressPercent != 0 {
		t.Errorf("shelf book = %+v", got)
	}
}

func TestShelfBooksAdd(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		addFirst   bool
		wantStatus int
	}{
		{"success", "/api/shelves/1/books", `{"book_id":1}`, false, http.StatusCreated},
		{"already on shelf", "/api/shelves/1/books", `{"book_id":1}`, true, http.StatusConflict},
		{"invalid shelf id", "/api/shelves/abc/books", `{"book_id":1}`, false, http.StatusBadRequest},
		{"invalid body", "/api/shelves/1/books", `{nope`, false, http.StatusBadRequest},
		{"unknown field", "/api/shelves/1/books", `{"book_id":1,"x":2}`, false, http.StatusBadRequest},
		{"missing book id", "/api/shelves/1/books", `{}`, false, http.StatusBadRequest},
		{"unknown shelf", "/api/shelves/99/books", `{"book_id":1}`, false, http.StatusNotFound},
		{"unknown book", "/api/shelves/1/books", `{"book_id":99}`, false, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newShelfBooksServer(t)
			if tt.addFirst {
				if rec := doRequest(s, http.MethodPost, tt.path, tt.body); rec.Code != http.StatusCreated {
					t.Fatalf("setup status = %d", rec.Code)
				}
			}
			rec := doRequest(s, http.MethodPost, tt.path, tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus != http.StatusCreated {
				return
			}
			assertNewShelfBook(t, rec.Body.Bytes())
		})
	}
}

func TestShelfBooksList(t *testing.T) {
	s := newShelfBooksServer(t)

	rec := doRequest(s, http.MethodGet, "/api/shelves/1/books", "")
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("empty list: %d %q", rec.Code, rec.Body.String())
	}

	doRequest(s, http.MethodPost, "/api/shelves/1/books", `{"book_id":1}`)
	rec = doRequest(s, http.MethodGet, "/api/shelves/1/books", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var got []db.ListBooksOnShelfRow
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].BookID != 1 || got[0].Title != "Title" || got[0].ProgressPercent != 0 {
		t.Errorf("list = %+v", got)
	}
}

func TestShelfBooksListErrors(t *testing.T) {
	s := newShelfBooksServer(t)
	for path, want := range map[string]int{
		"/api/shelves/abc/books": http.StatusBadRequest,
		"/api/shelves/99/books":  http.StatusNotFound,
	} {
		if rec := doRequest(s, http.MethodGet, path, ""); rec.Code != want {
			t.Errorf("%s: status = %d, want %d", path, rec.Code, want)
		}
	}
}

func TestShelfBooksRemove(t *testing.T) {
	s := newShelfBooksServer(t)
	doRequest(s, http.MethodPost, "/api/shelves/1/books", `{"book_id":1}`)

	tests := []struct {
		name, path string
		want       int
	}{
		{"invalid shelf id", "/api/shelves/abc/books/1", http.StatusBadRequest},
		{"invalid book id", "/api/shelves/1/books/abc", http.StatusBadRequest},
		{"unknown shelf", "/api/shelves/99/books/1", http.StatusNotFound},
		{"book not on shelf", "/api/shelves/1/books/99", http.StatusNotFound},
		{"success", "/api/shelves/1/books/1", http.StatusNoContent},
		{"already removed", "/api/shelves/1/books/1", http.StatusNotFound},
	}
	for _, tt := range tests {
		if rec := doRequest(s, http.MethodDelete, tt.path, ""); rec.Code != tt.want {
			t.Errorf("%s: status = %d, want %d", tt.name, rec.Code, tt.want)
		}
	}

	rec := doRequest(s, http.MethodGet, "/api/shelves/1/books", "")
	if strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Errorf("list after remove = %q", rec.Body.String())
	}
}
