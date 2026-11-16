package server

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(old) })
	return &buf
}

func serve(h http.Handler, id string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	if id != "" {
		req.Header.Set("X-Request-ID", id)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func chain(h http.Handler) http.Handler {
	return requestIDMiddleware(loggingMiddleware(recoveryMiddleware(h)))
}

func TestRequestID(t *testing.T) {
	var ctxID string
	h := chain(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ctxID = RequestID(r.Context())
	}))

	t.Run("reuses valid", func(t *testing.T) {
		captureLogs(t)
		rr := serve(h, "abc.DEF_123-x")
		if got := rr.Header().Get("X-Request-ID"); got != "abc.DEF_123-x" || ctxID != got {
			t.Fatalf("header %q ctx %q", got, ctxID)
		}
	})

	t.Run("accepts 128 chars", func(t *testing.T) {
		captureLogs(t)
		id := strings.Repeat("a", 128)
		if got := serve(h, id).Header().Get("X-Request-ID"); got != id {
			t.Fatalf("got %q", got)
		}
	})

	bad := map[string]string{
		"too long":  strings.Repeat("a", 129),
		"space":     "a b",
		"newline":   "a\nb",
		"slash":     "a/b",
		"non-ascii": "idé",
	}
	for name, id := range bad {
		t.Run("rejects "+name, func(t *testing.T) {
			captureLogs(t)
			got := serve(h, id).Header().Get("X-Request-ID")
			if got == id || !validRequestID(got) {
				t.Fatalf("got %q", got)
			}
		})
	}

	t.Run("generates when absent", func(t *testing.T) {
		captureLogs(t)
		a := serve(h, "").Header().Get("X-Request-ID")
		b := serve(h, "").Header().Get("X-Request-ID")
		if a == "" || !validRequestID(a) || a == b {
			t.Fatalf("a=%q b=%q", a, b)
		}
	})
}

func TestAccessLog(t *testing.T) {
	t.Run("implicit 200", func(t *testing.T) {
		buf := captureLogs(t)
		serve(chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})), "rid-1")
		out := buf.String()
		for _, want := range []string{"method=GET", "path=/x", "status=200", "duration=", "request_id=rid-1"} {
			if !strings.Contains(out, want) {
				t.Errorf("log %q missing %q", out, want)
			}
		}
	})

	t.Run("explicit status", func(t *testing.T) {
		buf := captureLogs(t)
		serve(chain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})), "rid-2")
		if !strings.Contains(buf.String(), "status=418") {
			t.Errorf("log %q", buf.String())
		}
	})

	t.Run("panic logs 500 and request ID", func(t *testing.T) {
		buf := captureLogs(t)
		rr := serve(chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("boom")
		})), "rid-3")
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("code %d", rr.Code)
		}
		if rr.Header().Get("X-Request-ID") != "rid-3" {
			t.Errorf("missing header on panic response")
		}
		out := buf.String()
		if !strings.Contains(out, "status=500") || strings.Count(out, "request_id=rid-3") != 2 {
			t.Errorf("log %q", out)
		}
	})
}

func TestHandlerSetsRequestID(t *testing.T) {
	captureLogs(t)
	s, err := New(nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := serve(s.Handler(), "")
	if rr.Header().Get("X-Request-ID") == "" {
		t.Fatal("no X-Request-ID on response")
	}
}
