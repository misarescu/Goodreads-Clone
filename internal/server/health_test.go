package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	s.handleHealth(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"status":"ok"}` {
		t.Errorf("body = %q, want %q", got, `{"status":"ok"}`)
	}
}

func TestHandleReady(t *testing.T) {
	s := newShelfBooksServer(t)
	rec := doRequest(s, http.MethodGet, "/api/ready", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"status":"ready"}` {
		t.Errorf("body = %q", got)
	}
}

func TestHandleReadyDatabaseClosed(t *testing.T) {
	s := newShelfBooksServer(t)
	if err := s.business.Storage.DB().Close(); err != nil {
		t.Fatal(err)
	}
	rec := doRequest(s, http.MethodGet, "/api/ready", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"status":"unavailable"}` {
		t.Errorf("body = %q", got)
	}
	// Liveness is unaffected.
	if rec := doRequest(s, http.MethodGet, "/api/health", ""); rec.Code != http.StatusOK {
		t.Errorf("health status = %d, want 200", rec.Code)
	}
}
