package server

import "net/http"

// handleHealth reports that the server is up. It does not touch the database.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
