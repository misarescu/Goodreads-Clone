package server

import (
	"net/http"

	"goodreads-clone.com/internal/business"
)

type Server struct {
	business *business.Business
	mux      *http.ServeMux
}

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

func (s *Server) Handler() http.Handler {
	return recoveryMiddleware(loggingMiddleware(s))
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}
