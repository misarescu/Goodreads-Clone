package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"goodreads-clone.com/internal/business"
)

// Server handles HTTP requests and routes them to the appropriate business logic.
type Server struct {
	business *business.Business
	mux      *http.ServeMux
}

// New creates a new Server instance with the given business layer.
func New(b *business.Business) (*Server, error) {
	s := &Server{
		business: b,
		mux:      http.NewServeMux(),
	}

	s.registerRoutes()

	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Handler returns the HTTP handler with middleware applied.
func (s *Server) Handler() http.Handler {
	return recoveryMiddleware(loggingMiddleware(s))
}

// ListenAndServe listens on addr and serves until ctx is cancelled, then shuts
// down gracefully. See Serve.
func (s *Server) ListenAndServe(ctx context.Context, addr string, shutdownTimeout time.Duration) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	return s.Serve(ctx, ln, shutdownTimeout)
}

// Serve serves on ln until ctx is cancelled. It then stops accepting new
// connections and waits up to shutdownTimeout for in-flight requests to finish.
// It returns nil after a clean shutdown.
func (s *Server) Serve(ctx context.Context, ln net.Listener, shutdownTimeout time.Duration) error {
	httpSrv := &http.Server{
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() { errCh <- httpSrv.Serve(ln) }()

	select {
	case err := <-errCh:
		return fmt.Errorf("HTTP server error: %w", err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		_ = httpSrv.Close()
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("HTTP server error: %w", err)
	}
	return nil
}
