package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func execAll(t *testing.T, s *Server, queries ...string) {
	t.Helper()
	for _, q := range queries {
		if _, err := s.business.Storage.DB().Exec(q); err != nil {
			t.Fatalf("%q: %v", q, err)
		}
	}
}

// newRecommendationsServer seeds, on top of newShelfBooksServer's user 1,
// shelf 1 and book 1 ("Title"):
//
//	users 1 (the reader), 2..4 (similar), 5 (not similar; finished Other, Gamma),
//	6 (only has book 1 on a shelf, unfinished), 7 (finished nothing)
//	books 2 Mine2, 3 Shelved, 4 Alpha, 5 Beta, 6 Gamma, 7 Delta, 8 Other, 9 Nine
func newRecommendationsServer(t *testing.T) *Server {
	t.Helper()
	s := newShelfBooksServer(t)
	const done = `'2026-01-01 00:00:00'`
	execAll(t, s,
		`INSERT INTO users (email, username, password_hash) VALUES
			('2@x', 'u2', 'x'), ('3@x', 'u3', 'x'), ('4@x', 'u4', 'x'),
			('5@x', 'u5', 'x'), ('6@x', 'u6', 'x'), ('7@x', 'u7', 'x')`,
		`INSERT INTO shelves (user_id, name) VALUES (2, 'S'), (3, 'S'), (4, 'S'), (5, 'S'), (6, 'S'), (7, 'S')`,
		`INSERT INTO books (title, author) VALUES
			('Mine2', 'A'), ('Shelved', 'A'), ('Alpha', 'A1'), ('Beta', 'B1'),
			('Gamma', 'C1'), ('Delta', 'D1'), ('Other', 'A'), ('Nine', 'A')`,
		// Shelf ids: user 1 -> 1, users 2..7 -> 2..7.
		// User 1 finished books 1, 2 and has book 3 on a shelf, unfinished.
		`INSERT INTO shelf_books (shelf_id, book_id, finished_at) VALUES
			(1, 1, `+done+`), (1, 2, `+done+`), (1, 3, NULL)`,
		// User 2 finished 1,2,3,4,5; user 3: 1,4,6; user 4: 2,4,7.
		`INSERT INTO shelf_books (shelf_id, book_id, finished_at) VALUES
			(2, 1, `+done+`), (2, 2, `+done+`), (2, 3, `+done+`), (2, 4, `+done+`), (2, 5, `+done+`),
			(3, 1, `+done+`), (3, 4, `+done+`), (3, 6, `+done+`),
			(4, 2, `+done+`), (4, 4, `+done+`), (4, 7, `+done+`)`,
		// User 5 finished Other and Gamma: not similar to user 1, but a Gamma finisher.
		`INSERT INTO shelf_books (shelf_id, book_id, finished_at) VALUES (5, 8, `+done+`), (5, 6, `+done+`)`,
		// User 6 only has the shared book 1 unfinished, and finished Nine.
		`INSERT INTO shelf_books (shelf_id, book_id, finished_at) VALUES (6, 1, NULL), (6, 9, `+done+`)`,
	)
	return s
}

type recResult struct {
	ID      int64    `json:"id"`
	Title   string   `json:"title"`
	Author  string   `json:"author"`
	Score   int64    `json:"score"`
	Because []string `json:"because"`
}

func getRecommendations(t *testing.T, s *Server, path string) []recResult {
	t.Helper()
	rec := doRequest(s, http.MethodGet, path, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s = %d: %s", path, rec.Code, rec.Body)
	}
	var out []recResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode %s: %v", rec.Body, err)
	}
	return out
}

func TestRecommendations(t *testing.T) {
	s := newRecommendationsServer(t)
	got := getRecommendations(t, s, "/api/users/1/recommendations")

	// Alpha: 3 similar readers. Gamma beats Beta and Delta on overall
	// finishers (user 5 also finished it); Beta before Delta by title.
	// Shelved (on user 1's shelf), Other and Nine (not similar readers) are out.
	want := []recResult{
		{4, "Alpha", "A1", 3, []string{"Mine2", "Title"}},
		{6, "Gamma", "C1", 1, []string{"Title"}},
		{5, "Beta", "B1", 1, []string{"Mine2", "Title"}},
		{7, "Delta", "D1", 1, []string{"Mine2"}},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i].ID != want[i].ID || got[i].Title != want[i].Title || got[i].Author != want[i].Author ||
			got[i].Score != want[i].Score || strings.Join(got[i].Because, "|") != strings.Join(want[i].Because, "|") {
			t.Errorf("result %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestRecommendationsExcludeShelvedUnfinished(t *testing.T) {
	// A candidate that is on one of the user's shelves, even unfinished, is excluded.
	s := newRecommendationsServer(t)
	execAll(t, s, `INSERT INTO shelves (user_id, name) VALUES (1, 'Later')`,
		`INSERT INTO shelf_books (shelf_id, book_id) VALUES (8, 4)`)
	for _, r := range getRecommendations(t, s, "/api/users/1/recommendations") {
		if r.Title == "Alpha" {
			t.Errorf("shelved book recommended: %+v", r)
		}
	}
}

func TestRecommendationsBecauseCappedAtThree(t *testing.T) {
	s := newRecommendationsServer(t)
	execAll(t, s,
		`INSERT INTO books (title, author) VALUES ('Aaa', 'A'), ('Bbb', 'A')`,
		`INSERT INTO shelf_books (shelf_id, book_id, finished_at) VALUES
			(1, 10, '2026-01-01'), (1, 11, '2026-01-01'),
			(2, 10, '2026-01-01'), (2, 11, '2026-01-01')`)
	got := getRecommendations(t, s, "/api/users/1/recommendations")
	if len(got[0].Because) != 3 || strings.Join(got[0].Because, "|") != "Aaa|Bbb|Mine2" {
		t.Errorf("because = %v", got[0].Because)
	}
}

func TestRecommendationsEmpty(t *testing.T) {
	s := newRecommendationsServer(t)
	for _, id := range []string{"6", "7"} {
		rec := doRequest(s, http.MethodGet, "/api/users/"+id+"/recommendations", "")
		if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "[]" {
			t.Errorf("user %s: %d %q, want 200 []", id, rec.Code, rec.Body)
		}
	}
}

func TestRecommendationsLimit(t *testing.T) {
	s := newRecommendationsServer(t)
	execAll(t, s, `WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i + 1 FROM n WHERE i < 60)
		INSERT INTO books (title, author) SELECT 'Bulk ' || i, 'A' FROM n`,
		`INSERT INTO shelf_books (shelf_id, book_id, finished_at)
			SELECT 2, id, '2026-01-01' FROM books WHERE title LIKE 'Bulk %'`)

	tests := []struct {
		query string
		want  int
	}{
		{"", 10},
		{"?limit=2", 2},
		{"?limit=50", 50},
		{"?limit=1000", 50},
	}
	for _, tt := range tests {
		if got := getRecommendations(t, s, "/api/users/1/recommendations"+tt.query); len(got) != tt.want {
			t.Errorf("%q: %d results, want %d", tt.query, len(got), tt.want)
		}
	}
	got := getRecommendations(t, s, "/api/users/1/recommendations?limit=2")
	if got[0].Title != "Alpha" {
		t.Errorf("top = %q, want Alpha", got[0].Title)
	}
}

func TestRecommendationsErrors(t *testing.T) {
	s := newRecommendationsServer(t)
	for path, want := range map[string]int{
		"/api/users/99/recommendations":          http.StatusNotFound,
		"/api/users/x/recommendations":           http.StatusBadRequest,
		"/api/users/1/recommendations?limit=0":   http.StatusBadRequest,
		"/api/users/1/recommendations?limit=-1":  http.StatusBadRequest,
		"/api/users/1/recommendations?limit=abc": http.StatusBadRequest,
	} {
		if rec := doRequest(s, http.MethodGet, path, ""); rec.Code != want {
			t.Errorf("%s = %d, want %d", path, rec.Code, want)
		}
	}
}
