package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func seedReviewUsers(t *testing.T, s *Server) {
	t.Helper()
	for _, q := range []string{
		`INSERT INTO users (email, username, password_hash) VALUES ('b@b.c', 'bob', 'x')`,
		`INSERT INTO books (title, author) VALUES ('B2', 'A')`,
	} {
		if _, err := s.business.Storage.DB().Exec(q); err != nil {
			t.Fatal(err)
		}
	}
}

func TestReviewSave(t *testing.T) {
	long := strings.Repeat("é", 5000)
	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
	}{
		{"create", "/api/books/1/reviews", `{"user_id":1,"rating":4,"body":"good"}`, http.StatusCreated},
		{"max body", "/api/books/1/reviews", `{"user_id":1,"rating":5,"body":"` + long + `"}`, http.StatusCreated},
		{"body too long", "/api/books/1/reviews", `{"user_id":1,"rating":5,"body":"` + long + `x"}`, http.StatusBadRequest},
		{"empty body", "/api/books/1/reviews", `{"user_id":1,"rating":5,"body":""}`, http.StatusBadRequest},
		{"rating 0", "/api/books/1/reviews", `{"user_id":1,"rating":0,"body":"x"}`, http.StatusBadRequest},
		{"rating 6", "/api/books/1/reviews", `{"user_id":1,"rating":6,"body":"x"}`, http.StatusBadRequest},
		{"missing rating", "/api/books/1/reviews", `{"user_id":1,"body":"x"}`, http.StatusBadRequest},
		{"missing user", "/api/books/1/reviews", `{"rating":3,"body":"x"}`, http.StatusBadRequest},
		{"bad json", "/api/books/1/reviews", `{nope`, http.StatusBadRequest},
		{"bad book id", "/api/books/x/reviews", `{"user_id":1,"rating":3,"body":"x"}`, http.StatusBadRequest},
		{"unknown book", "/api/books/99/reviews", `{"user_id":1,"rating":3,"body":"x"}`, http.StatusNotFound},
		{"unknown user", "/api/books/1/reviews", `{"user_id":99,"rating":3,"body":"x"}`, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newShelfBooksServer(t)
			rec := doRequest(s, http.MethodPost, tt.path, tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tt.wantStatus, rec.Body)
			}
		})
	}
}

func TestReviewReplace(t *testing.T) {
	s := newShelfBooksServer(t)
	rec := doRequest(s, http.MethodPost, "/api/books/1/reviews", `{"user_id":1,"rating":2,"body":"meh"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first status = %d", rec.Code)
	}
	rec = doRequest(s, http.MethodPost, "/api/books/1/reviews", `{"user_id":1,"rating":5,"body":"loved it"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("second status = %d: %s", rec.Code, rec.Body)
	}
	var got struct {
		Rating int64  `json:"rating"`
		Body   string `json:"body"`
		BookID int64  `json:"book_id"`
		UserID int64  `json:"user_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Rating != 5 || got.Body != "loved it" || got.BookID != 1 || got.UserID != 1 {
		t.Errorf("review = %+v", got)
	}
	var n int
	if err := s.business.Storage.DB().QueryRow(`SELECT COUNT(*) FROM reviews`).Scan(&n); err != nil || n != 1 {
		t.Errorf("rows = %d, err = %v", n, err)
	}
}

func seedReviewRows(t *testing.T, s *Server) {
	t.Helper()
	seedReviewUsers(t, s)
	_, err := s.business.Storage.DB().Exec(`INSERT INTO reviews (book_id, user_id, rating, body, created_at) VALUES
		(1, 1, 3, 'older', '2026-01-01 10:00:00'),
		(1, 2, 5, 'newer', '2026-02-01 10:00:00')`)
	if err != nil {
		t.Fatal(err)
	}
}

type listedReview struct {
	Username string `json:"username"`
	Body     string `json:"body"`
}

func listReviews(s *Server, path string) (int, []listedReview) {
	rec := doRequest(s, http.MethodGet, path, "")
	var rows []listedReview
	_ = json.Unmarshal(rec.Body.Bytes(), &rows)
	return rec.Code, rows
}

func TestReviewListOrderAndPaging(t *testing.T) {
	s := newShelfBooksServer(t)
	seedReviewRows(t, s)

	code, rows := listReviews(s, "/api/books/1/reviews")
	want := []listedReview{{"bob", "newer"}, {"u", "older"}}
	if code != http.StatusOK || !reflect.DeepEqual(rows, want) {
		t.Errorf("list = %d %+v", code, rows)
	}
	_, rows = listReviews(s, "/api/books/1/reviews?limit=1&offset=1")
	if !reflect.DeepEqual(rows, want[1:]) {
		t.Errorf("page = %+v", rows)
	}
}

func TestReviewListEdges(t *testing.T) {
	s := newShelfBooksServer(t)
	seedReviewRows(t, s)

	code, rows := listReviews(s, "/api/books/2/reviews")
	if code != http.StatusOK || rows == nil || len(rows) != 0 {
		t.Errorf("empty = %d %+v", code, rows)
	}
	if code, _ = listReviews(s, "/api/books/99/reviews"); code != http.StatusNotFound {
		t.Errorf("unknown book = %d", code)
	}
	if code, _ = listReviews(s, "/api/books/x/reviews"); code != http.StatusBadRequest {
		t.Errorf("bad id = %d", code)
	}
}

func TestReviewDelete(t *testing.T) {
	s := newShelfBooksServer(t)
	doRequest(s, http.MethodPost, "/api/books/1/reviews", `{"user_id":1,"rating":4,"body":"x"}`)
	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"bad book", "/api/books/x/reviews/1", http.StatusBadRequest},
		{"bad user", "/api/books/1/reviews/x", http.StatusBadRequest},
		{"no such review", "/api/books/1/reviews/99", http.StatusNotFound},
		{"delete", "/api/books/1/reviews/1", http.StatusNoContent},
		{"delete again", "/api/books/1/reviews/1", http.StatusNotFound},
	}
	for _, tt := range tests {
		if rec := doRequest(s, http.MethodDelete, tt.path, ""); rec.Code != tt.wantStatus {
			t.Errorf("%s: status = %d, want %d", tt.name, rec.Code, tt.wantStatus)
		}
	}
}

func TestReviewCascades(t *testing.T) {
	s := newShelfBooksServer(t)
	doRequest(s, http.MethodPost, "/api/books/1/reviews", `{"user_id":1,"rating":4,"body":"x"}`)
	if rec := doRequest(s, http.MethodDelete, "/api/books/1", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("delete book = %d", rec.Code)
	}
	var n int
	if err := s.business.Storage.DB().QueryRow(`SELECT COUNT(*) FROM reviews`).Scan(&n); err != nil || n != 0 {
		t.Errorf("rows = %d, err = %v", n, err)
	}
}

type bookStats struct {
	ID            int64    `json:"id"`
	AverageRating *float64 `json:"average_rating"`
	ReviewCount   int64    `json:"review_count"`
}

func getJSON(t *testing.T, s *Server, path string, v any) {
	t.Helper()
	rec := doRequest(s, http.MethodGet, path, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("%s = %d", path, rec.Code)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
		t.Fatal(err)
	}
}

func checkStats(t *testing.T, got bookStats, avg float64, count int64) {
	t.Helper()
	if got.AverageRating == nil || *got.AverageRating != avg || got.ReviewCount != count {
		t.Errorf("stats = %+v, want avg %v count %d", got, avg, count)
	}
}

func postReview(s *Server, bookID, userID, rating int) {
	body := fmt.Sprintf(`{"user_id":%d,"rating":%d,"body":"b"}`, userID, rating)
	doRequest(s, http.MethodPost, fmt.Sprintf("/api/books/%d/reviews", bookID), body)
}

func TestBookReviewStatsWithoutReviews(t *testing.T) {
	s := newShelfBooksServer(t)
	var b bookStats
	getJSON(t, s, "/api/books/1", &b)
	if b.AverageRating != nil || b.ReviewCount != 0 {
		t.Errorf("no reviews: %+v", b)
	}
	if !strings.Contains(doRequest(s, http.MethodGet, "/api/books/1", "").Body.String(), `"average_rating":null`) {
		t.Error("average_rating should be null without reviews")
	}
}

func TestBookReviewStats(t *testing.T) {
	s := newShelfBooksServer(t)
	seedReviewUsers(t, s)
	postReview(s, 1, 1, 4)
	postReview(s, 1, 2, 3)
	postReview(s, 1, 2, 4) // replaces bob's 3
	postReview(s, 2, 1, 4)
	postReview(s, 2, 2, 3)

	var b bookStats
	getJSON(t, s, "/api/books/1", &b)
	checkStats(t, b, 4, 2)

	var list []bookStats
	getJSON(t, s, "/api/books", &list)
	if len(list) != 2 {
		t.Fatalf("list = %+v", list)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	checkStats(t, list[1], 3.5, 2)
	checkStats(t, list[0], 4, 2)

	var byISBN bookStats
	if _, err := s.business.Storage.DB().Exec(`UPDATE books SET isbn = 'i1' WHERE id = 1`); err != nil {
		t.Fatal(err)
	}
	getJSON(t, s, "/api/books?isbn=i1", &byISBN)
	checkStats(t, byISBN, 4, 2)
}

func TestBookReviewStatsRounds(t *testing.T) {
	s := newShelfBooksServer(t)
	seedReviewUsers(t, s)
	if _, err := s.business.Storage.DB().Exec(
		`INSERT INTO users (email, username, password_hash) VALUES ('d@b.c', 'dd', 'x')`); err != nil {
		t.Fatal(err)
	}
	postReview(s, 1, 1, 4)
	postReview(s, 1, 2, 5)
	postReview(s, 1, 3, 4) // 13/3 = 4.333...

	var b bookStats
	getJSON(t, s, "/api/books/1", &b)
	checkStats(t, b, 4.3, 3)
}

type userReviewItem struct {
	BookTitle string `json:"book_title"`
	Body      string `json:"body"`
}

func getUserReviews(s *Server, path string) (int, []userReviewItem) {
	rec := doRequest(s, http.MethodGet, path, "")
	var rows []userReviewItem
	_ = json.Unmarshal(rec.Body.Bytes(), &rows)
	return rec.Code, rows
}

func TestUserReviewList(t *testing.T) {
	s := newShelfBooksServer(t)
	seedReviewRows(t, s)
	// user 1 also reviews book 2, most recently.
	if _, err := s.business.Storage.DB().Exec(`INSERT INTO reviews (book_id, user_id, rating, body, created_at)
		VALUES (2, 1, 4, 'latest', '2026-03-01 10:00:00')`); err != nil {
		t.Fatal(err)
	}

	code, rows := getUserReviews(s, "/api/users/1/reviews")
	want := []userReviewItem{{"B2", "latest"}, {"Title", "older"}}
	if code != http.StatusOK || !reflect.DeepEqual(rows, want) {
		t.Errorf("list = %d %+v", code, rows)
	}
	if _, rows = getUserReviews(s, "/api/users/1/reviews?limit=1&offset=1"); !reflect.DeepEqual(rows, want[1:]) {
		t.Errorf("page = %+v", rows)
	}
}

func TestUserReviewListEdges(t *testing.T) {
	s := newShelfBooksServer(t)
	seedReviewRows(t, s)

	for path, want := range map[string]int{
		"/api/users/2/reviews":  http.StatusOK,
		"/api/users/99/reviews": http.StatusNotFound,
		"/api/users/x/reviews":  http.StatusBadRequest,
	} {
		if code, _ := getUserReviews(s, path); code != want {
			t.Errorf("%s = %d, want %d", path, code, want)
		}
	}
}

func seedSortRows(t *testing.T, s *Server) {
	t.Helper()
	seedReviewUsers(t, s)
	const addUser = `INSERT INTO users (email, username, password_hash) VALUES ('c@c.c', 'cat', 'x')`
	if _, err := s.business.Storage.DB().Exec(addUser); err != nil {
		t.Fatal(err)
	}
	_, err := s.business.Storage.DB().Exec(`INSERT INTO reviews (book_id, user_id, rating, body, created_at) VALUES
		(1, 1, 3, 'three', '2026-01-01 10:00:00'),
		(1, 2, 5, 'five-old', '2026-02-01 10:00:00'),
		(1, 3, 5, 'five-new', '2026-03-01 10:00:00')`)
	if err != nil {
		t.Fatal(err)
	}
}

func bodies(rows []listedReview) []string {
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Body)
	}
	return out
}

func TestReviewListSortAndFilter(t *testing.T) {
	s := newShelfBooksServer(t)
	seedSortRows(t, s)

	for _, tc := range []struct {
		query string
		want  []string
	}{
		{"", []string{"five-new", "five-old", "three"}},
		{"?sort=recent", []string{"five-new", "five-old", "three"}},
		{"?sort=rating", []string{"five-new", "five-old", "three"}},
		{"?min_rating=4", []string{"five-new", "five-old"}},
		{"?min_rating=1", []string{"five-new", "five-old", "three"}},
		{"?min_rating=5&sort=rating", []string{"five-new", "five-old"}},
		{"?sort=rating&limit=2", []string{"five-new", "five-old"}},
		{"?sort=rating&limit=2&offset=2", []string{"three"}},
		{"?sort=rating&min_rating=3&limit=1&offset=1", []string{"five-old"}},
		{"?sort=recent&min_rating=4&limit=1&offset=1", []string{"five-old"}},
	} {
		code, rows := listReviews(s, "/api/books/1/reviews"+tc.query)
		if code != http.StatusOK || !reflect.DeepEqual(bodies(rows), tc.want) {
			t.Errorf("%q = %d %v, want %v", tc.query, code, bodies(rows), tc.want)
		}
	}
}

func TestReviewListSortByRatingOrdersByRatingFirst(t *testing.T) {
	s := newShelfBooksServer(t)
	seedReviewUsers(t, s)
	_, err := s.business.Storage.DB().Exec(`INSERT INTO reviews (book_id, user_id, rating, body, created_at) VALUES
		(1, 1, 5, 'high-old', '2026-01-01 10:00:00'),
		(1, 2, 2, 'low-new', '2026-02-01 10:00:00')`)
	if err != nil {
		t.Fatal(err)
	}
	_, rows := listReviews(s, "/api/books/1/reviews?sort=rating")
	if want := []string{"high-old", "low-new"}; !reflect.DeepEqual(bodies(rows), want) {
		t.Errorf("rating order = %v", bodies(rows))
	}
	_, rows = listReviews(s, "/api/books/1/reviews")
	if want := []string{"low-new", "high-old"}; !reflect.DeepEqual(bodies(rows), want) {
		t.Errorf("recent order = %v", bodies(rows))
	}
}

func TestReviewListBadParams(t *testing.T) {
	s := newShelfBooksServer(t)
	seedSortRows(t, s)

	for _, tc := range []struct{ query, param string }{
		{"?sort=best", "sort"},
		{"?sort=", ""},
		{"?min_rating=0", "min_rating"},
		{"?min_rating=6", "min_rating"},
		{"?min_rating=-1", "min_rating"},
		{"?min_rating=abc", "min_rating"},
		{"?min_rating=2.5", "min_rating"},
	} {
		rec := doRequest(s, http.MethodGet, "/api/books/1/reviews"+tc.query, "")
		if tc.param == "" {
			if rec.Code != http.StatusOK {
				t.Errorf("%q = %d", tc.query, rec.Code)
			}
			continue
		}
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), tc.param) {
			t.Errorf("%q = %d %s", tc.query, rec.Code, rec.Body)
		}
	}
}
