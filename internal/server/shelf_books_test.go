package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

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

func TestShelfBooksProgress(t *testing.T) {
	s := newShelfBooksServer(t)
	// Book 1 has no page count; book 2 has 300 pages.
	if _, err := s.business.CreateBook(context.Background(), db.CreateBookParams{
		Title: "Long", Author: "A", PageCount: sql.NullInt64{Int64: 300, Valid: true},
	}); err != nil {
		t.Fatal(err)
	}
	addShelfBooks(t, s, "1", "2")

	tests := []struct {
		name         string
		path, body   string
		wantStatus   int
		wantPages    int64
		wantPercent  int64
		wantFinished bool
	}{
		{"pages computes percent rounded down", "/api/shelves/1/books/2/progress", `{"pages_read":100}`, 200, 100, 33, false},
		{"reaching the page count finishes", "/api/shelves/1/books/2/progress", `{"pages_read":300}`, 200, 300, 100, true},
		{"staying finished keeps it", "/api/shelves/1/books/2/progress", `{"pages_read":300}`, 200, 300, 100, true},
		{"dropping below 100 clears", "/api/shelves/1/books/2/progress", `{"pages_read":299}`, 200, 299, 99, false},
		{"above page count", "/api/shelves/1/books/2/progress", `{"pages_read":301}`, 400, 0, 0, false},
		{"negative pages", "/api/shelves/1/books/2/progress", `{"pages_read":-1}`, 400, 0, 0, false},
		{"percent for book with page count", "/api/shelves/1/books/2/progress", `{"progress_percent":40}`, 400, 0, 0, false},
		{"percent without page count", "/api/shelves/1/books/1/progress", `{"progress_percent":40}`, 200, 0, 40, false},
		{"percent 100 finishes", "/api/shelves/1/books/1/progress", `{"progress_percent":100}`, 200, 0, 100, true},
		{"percent above 100", "/api/shelves/1/books/1/progress", `{"progress_percent":101}`, 400, 0, 0, false},
		{"pages without page count", "/api/shelves/1/books/1/progress", `{"pages_read":5}`, 400, 0, 0, false},
		{"empty body", "/api/shelves/1/books/1/progress", `{}`, 400, 0, 0, false},
		{"unknown field", "/api/shelves/1/books/1/progress", `{"x":1}`, 400, 0, 0, false},
		{"invalid shelf id", "/api/shelves/abc/books/1/progress", `{"pages_read":1}`, 400, 0, 0, false},
		{"invalid book id", "/api/shelves/1/books/abc/progress", `{"pages_read":1}`, 400, 0, 0, false},
		{"book not on shelf", "/api/shelves/1/books/99/progress", `{"pages_read":1}`, 404, 0, 0, false},
	}
	for _, tt := range tests {
		rec := doRequest(s, http.MethodPut, tt.path, tt.body)
		if rec.Code != tt.wantStatus {
			t.Fatalf("%s: status = %d, want %d (%s)", tt.name, rec.Code, tt.wantStatus, rec.Body.String())
		}
		if tt.wantStatus != http.StatusOK {
			continue
		}
		var got map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got["pages_read"] != float64(tt.wantPages) || got["progress_percent"] != float64(tt.wantPercent) {
			t.Errorf("%s: got %v", tt.name, got)
		}
		assertFinishedAt(t, tt.name, got, tt.wantFinished)
	}

	assertListFinishedAt(t, s, 1, true)
	assertListFinishedAt(t, s, 2, false)
}

func addShelfBooks(t *testing.T, s *Server, ids ...string) {
	t.Helper()
	for _, id := range ids {
		rec := doRequest(s, http.MethodPost, "/api/shelves/1/books", `{"book_id":`+id+`}`)
		if rec.Code != http.StatusCreated {
			t.Fatalf("add status = %d", rec.Code)
		}
	}
}

// assertFinishedAt checks finished_at is an RFC3339 string when finished, else JSON null.
func assertFinishedAt(t *testing.T, name string, b map[string]any, finished bool) {
	t.Helper()
	v, ok := b["finished_at"]
	if !ok {
		t.Errorf("%s: finished_at missing: %v", name, b)
		return
	}
	if !finished {
		if v != nil {
			t.Errorf("%s: finished_at = %v, want null", name, v)
		}
		return
	}
	str, isStr := v.(string)
	if !isStr || str == "" {
		t.Errorf("%s: finished_at = %v, want timestamp string", name, v)
		return
	}
	if _, err := time.Parse(time.RFC3339, str); err != nil {
		t.Errorf("%s: finished_at %q not RFC3339: %v", name, str, err)
	}
}

// assertListFinishedAt checks finished_at in the list response for a book.
func assertListFinishedAt(t *testing.T, s *Server, bookID int, finished bool) {
	t.Helper()
	rec := doRequest(s, http.MethodGet, "/api/shelves/1/books", "")
	var list []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	for _, b := range list {
		if b["book_id"] == float64(bookID) {
			assertFinishedAt(t, "list", b, finished)
			return
		}
	}
	t.Errorf("book %d missing from list: %s", bookID, rec.Body.String())
}

// seedMoveFixtures adds shelf 2 (same user), shelf 3 (another user) and a
// finished book entry (80/100 pages... recorded as 100%) for book 1 on shelf 1.
func seedMoveFixtures(t *testing.T, s *Server) *sql.DB {
	t.Helper()
	database := s.business.Storage.DB()
	for _, q := range []string{
		`INSERT INTO users (email, username, password_hash) VALUES ('b@b.c', 'u2', 'x')`,
		`INSERT INTO shelves (user_id, name) VALUES (1, 'Second')`,
		`INSERT INTO shelves (user_id, name) VALUES (2, 'Other user')`,
		`INSERT INTO shelf_books (shelf_id, book_id, pages_read, progress_percent, finished_at)
		 VALUES (1, 1, 120, 100, '2026-01-02 10:00:00')`,
	} {
		if _, err := database.Exec(q); err != nil {
			t.Fatalf("seed %q: %v", q, err)
		}
	}
	return database
}

func countOnShelf(t *testing.T, database *sql.DB, shelfID int) int {
	t.Helper()
	var n int
	if err := database.QueryRow(`SELECT COUNT(*) FROM shelf_books WHERE shelf_id = ?`, shelfID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestShelfBooksMoveKeepsProgress(t *testing.T) {
	s := newShelfBooksServer(t)
	database := seedMoveFixtures(t, s)

	rec := doRequest(s, http.MethodPost, "/api/shelves/1/books/1/move", `{"to_shelf_id":2}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d (%s)", rec.Code, rec.Body.String())
	}
	var got db.ListBooksOnShelfRow
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	assertMovedEntry(t, got)
	if countOnShelf(t, database, 1) != 0 || countOnShelf(t, database, 2) != 1 {
		t.Error("book should be only on shelf 2")
	}

	list := doRequest(s, http.MethodGet, "/api/shelves/2/books", "")
	var rows []db.ListBooksOnShelfRow
	if err := json.Unmarshal(list.Body.Bytes(), &rows); err != nil || len(rows) != 1 || rows[0].ID != got.ID {
		t.Errorf("list = %s (err %v)", list.Body.String(), err)
	}
}

func TestShelfBooksMoveErrors(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
	}{
		{"unknown source shelf", "/api/shelves/99/books/1/move", `{"to_shelf_id":2}`, http.StatusNotFound},
		{"unknown target shelf", "/api/shelves/1/books/1/move", `{"to_shelf_id":99}`, http.StatusNotFound},
		{"book not on source shelf", "/api/shelves/2/books/1/move", `{"to_shelf_id":1}`, http.StatusNotFound},
		{"same shelf", "/api/shelves/1/books/1/move", `{"to_shelf_id":1}`, http.StatusBadRequest},
		{"different users", "/api/shelves/1/books/1/move", `{"to_shelf_id":3}`, http.StatusBadRequest},
		{"missing target", "/api/shelves/1/books/1/move", `{}`, http.StatusBadRequest},
		{"invalid body", "/api/shelves/1/books/1/move", `{nope`, http.StatusBadRequest},
		{"invalid book id", "/api/shelves/1/books/x/move", `{"to_shelf_id":2}`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newShelfBooksServer(t)
			database := seedMoveFixtures(t, s)
			rec := doRequest(s, http.MethodPost, tt.path, tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if countOnShelf(t, database, 1) != 1 {
				t.Error("book must stay on shelf 1")
			}
		})
	}
}

func TestShelfBooksMoveConflict(t *testing.T) {
	t.Run("already on target shelf", func(t *testing.T) {
		s := newShelfBooksServer(t)
		database := seedMoveFixtures(t, s)
		if _, err := database.Exec(`INSERT INTO shelf_books (shelf_id, book_id, pages_read) VALUES (2, 1, 5)`); err != nil {
			t.Fatal(err)
		}
		rec := doRequest(s, http.MethodPost, "/api/shelves/1/books/1/move", `{"to_shelf_id":2}`)
		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d (%s)", rec.Code, rec.Body.String())
		}
		var pages1, pages2 int
		_ = database.QueryRow(`SELECT pages_read FROM shelf_books WHERE shelf_id = 1`).Scan(&pages1)
		_ = database.QueryRow(`SELECT pages_read FROM shelf_books WHERE shelf_id = 2`).Scan(&pages2)
		if pages1 != 120 || pages2 != 5 {
			t.Errorf("nothing should change: %d, %d", pages1, pages2)
		}
	})
}

func assertMovedEntry(t *testing.T, got db.ListBooksOnShelfRow) {
	t.Helper()
	want := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	if got.FinishedAt == nil || !got.FinishedAt.Equal(want) {
		t.Errorf("finished_at = %v, want %v", got.FinishedAt, want)
	}
}
