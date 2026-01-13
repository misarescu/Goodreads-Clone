package server

import (
	"net/http"

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

// ListenAndServe starts the HTTP server on the specified address.
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}
