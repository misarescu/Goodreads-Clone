package server

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestChallengePut(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
	}{
		{"create", "/api/users/1/challenges/2026", `{"goal":24}`, http.StatusOK},
		{"min", "/api/users/1/challenges/2026", `{"goal":1}`, http.StatusOK},
		{"max", "/api/users/1/challenges/2026", `{"goal":1000}`, http.StatusOK},
		{"zero", "/api/users/1/challenges/2026", `{"goal":0}`, http.StatusBadRequest},
		{"too big", "/api/users/1/challenges/2026", `{"goal":1001}`, http.StatusBadRequest},
		{"missing goal", "/api/users/1/challenges/2026", `{}`, http.StatusBadRequest},
		{"bad body", "/api/users/1/challenges/2026", `{nope`, http.StatusBadRequest},
		{"bad year", "/api/users/1/challenges/abc", `{"goal":5}`, http.StatusBadRequest},
		{"bad user", "/api/users/x/challenges/2026", `{"goal":5}`, http.StatusBadRequest},
		{"unknown user", "/api/users/99/challenges/2026", `{"goal":5}`, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newShelfBooksServer(t)
			rec := doRequest(s, http.MethodPut, tt.path, tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tt.wantStatus, rec.Body)
			}
		})
	}
}

func TestChallengePutUpdates(t *testing.T) {
	s := newShelfBooksServer(t)
	doRequest(s, http.MethodPut, "/api/users/1/challenges/2026", `{"goal":24}`)
	rec := doRequest(s, http.MethodPut, "/api/users/1/challenges/2026", `{"goal":30}`)
	if got := strings.TrimSpace(rec.Body.String()); got != `{"year":2026,"goal":30}` {
		t.Errorf("body = %s", got)
	}
	var n int
	if err := s.business.Storage.DB().QueryRow(`SELECT COUNT(*) FROM reading_challenges`).Scan(&n); err != nil || n != 1 {
		t.Errorf("rows = %d, err = %v", n, err)
	}
}

func TestChallengeGet(t *testing.T) {
	// 2026-07-02 12:00 UTC is exactly half of the 365-day year (182.5 days).
	s := newShelfBooksServer(t)
	s.business.Now = func() time.Time { return time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC) }

	exec := func(q string) {
		t.Helper()
		if _, err := s.business.Storage.DB().Exec(q); err != nil {
			t.Fatalf("%q: %v", q, err)
		}
	}
	exec(`INSERT INTO books (title, author) VALUES ('B2', 'A'), ('B3', 'A'), ('B4', 'A')`)
	exec(`INSERT INTO shelves (user_id, name) VALUES (1, 'Other')`)
	exec(`INSERT INTO shelf_books (shelf_id, book_id, finished_at) VALUES
		(1, 1, '2026-03-01 10:00:00'),
		(2, 1, '2026-03-02 10:00:00'),
		(1, 2, '2026-12-31 23:59:59'),
		(1, 3, '2025-12-31 23:59:59'),
		(1, 4, NULL)`)

	doRequest(s, http.MethodPut, "/api/users/1/challenges/2026", `{"goal":25}`)
	doRequest(s, http.MethodPut, "/api/users/1/challenges/2025", `{"goal":10}`)
	doRequest(s, http.MethodPut, "/api/users/1/challenges/2027", `{"goal":10}`)

	tests := []struct {
		name       string
		path       string
		wantStatus int
		want       string
	}{
		{"current year", "/api/users/1/challenges/2026", 200,
			`{"year":2026,"goal":25,"finished":2,"expected_by_now":12,"ahead_by":-10}`},
		{"past year", "/api/users/1/challenges/2025", 200,
			`{"year":2025,"goal":10,"finished":1,"expected_by_now":10,"ahead_by":-9}`},
		{"future year", "/api/users/1/challenges/2027", 200,
			`{"year":2027,"goal":10,"finished":0,"expected_by_now":0,"ahead_by":0}`},
		{"no challenge", "/api/users/1/challenges/2024", 404, ""},
		{"unknown user", "/api/users/99/challenges/2026", 404, ""},
		{"bad year", "/api/users/1/challenges/abc", 400, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(s, http.MethodGet, tt.path, "")
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tt.wantStatus, rec.Body)
			}
			if tt.want != "" && strings.TrimSpace(rec.Body.String()) != tt.want {
				t.Errorf("body = %s, want %s", rec.Body, tt.want)
			}
		})
	}
}
