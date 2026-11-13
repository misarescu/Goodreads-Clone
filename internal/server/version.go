package server

import "net/http"

// Version is the build version. It defaults to "dev" and can be set at build
// time with -ldflags "-X goodreads-clone.com/internal/server.Version=<version>".
var Version = "dev"

// handleVersion reports which build of the server is running.
func (s *Server) handleVersion(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=60")
	writeJSON(w, http.StatusOK, map[string]string{"version": Version})
}
