package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
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

func TestServesEmbeddedFrontend(t *testing.T) {
	const index = "<html>hi</html>"
	s := &Server{mux: http.NewServeMux(), staticFS: fstest.MapFS{
		"dist/index.html":    {Data: []byte(index)},
		"dist/assets/app.js": {Data: []byte("console.log(1)")},
	}}
	s.registerRoutes()

	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /: status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("GET /: Content-Type = %q", ct)
	}
	if rec.Body.String() != index {
		t.Errorf("GET /: body = %q", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "console.log(1)" {
		t.Errorf("GET /assets/app.js: status = %d, body = %q", rec.Code, rec.Body.String())
	}
}
