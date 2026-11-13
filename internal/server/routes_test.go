package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const contentTypeJSON = "application/" + "json"

func newRoutedServer() *Server {
	s := &Server{mux: http.NewServeMux()}
	s.registerRoutes()
	return s
}

func TestUnknownAPIRoute(t *testing.T) {
	s := newRoutedServer()
	for _, path := range []string{"/api/nope", "/api/", "/api/books/1/unknown"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		s.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: status = %d, want 404", path, rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != contentTypeJSON {
			t.Errorf("%s: Content-Type = %q", path, ct)
		}
		if got := strings.TrimSpace(rec.Body.String()); got != `{"error":"not found"}` {
			t.Errorf("%s: body = %q", path, got)
		}
	}
}

func TestExistingAPIRouteStillWorks(t *testing.T) {
	s := newRoutedServer()
	req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"version"`) {
		t.Errorf("body = %q", rec.Body.String())
	}
}
