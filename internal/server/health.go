package server

import "net/http"

// handleHealthz reports liveness for load balancers and Kubernetes probes.
// It requires no authentication and does not touch the database.
func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
