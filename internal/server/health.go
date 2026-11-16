package server

import "net/http"

// handleHealth reports that the server is up. It does not touch the database.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleReady reports whether the server can serve: 200 when the database
// answers a ping, 503 when it does not.
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if err := s.business.Ready(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
