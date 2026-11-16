package server

import (
	"io/fs"
	"net/http"

	"goodreads-clone.com/static"
)

func (s *Server) registerRoutes() {
	// Health API
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/ready", s.handleReady)
	s.mux.HandleFunc("GET /api/version", s.handleVersion)

	// Users API
	s.mux.HandleFunc("POST /api/users", s.handleCreateUser)
	s.mux.HandleFunc("GET /api/users/{id}", s.handleGetUserByID)
	s.mux.HandleFunc("GET /api/users/{id}/stats", s.handleGetUserStats)
	s.mux.HandleFunc("GET /api/users", s.handleGetUsers)
	s.mux.HandleFunc("PUT /api/users/{id}", s.handleUpdateUser)
	s.mux.HandleFunc("DELETE /api/users/{id}", s.handleDeleteUser)

	s.mux.HandleFunc("GET /api/users/{id}/recommendations", s.handleGetRecommendations)

	// Reading challenges API
	s.mux.HandleFunc("GET /api/users/{id}/challenges", s.handleListChallenges)
	s.mux.HandleFunc("PUT /api/users/{id}/challenges/{year}", s.handleSetChallenge)
	s.mux.HandleFunc("GET /api/users/{id}/challenges/{year}", s.handleGetChallenge)
	s.mux.HandleFunc("DELETE /api/users/{id}/challenges/{year}", s.handleDeleteChallenge)

	// Books API
	s.mux.HandleFunc("POST /api/books", s.handleCreateBook)
	s.mux.HandleFunc("GET /api/books/search", s.handleSearchBooks)
	s.mux.HandleFunc("GET /api/books/count", s.handleCountBooks)
	s.mux.HandleFunc("GET /api/books/{id}", s.handleGetBookByID)
	s.mux.HandleFunc("GET /api/books", s.handleGetBooks)
	s.mux.HandleFunc("PUT /api/books/{id}", s.handleUpdateBook)
	s.mux.HandleFunc("DELETE /api/books/{id}", s.handleDeleteBook)

	// Reviews API
	s.mux.HandleFunc("POST /api/books/{id}/reviews", s.handleSaveReview)
	s.mux.HandleFunc("GET /api/books/{id}/reviews", s.handleListReviews)
	s.mux.HandleFunc("DELETE /api/books/{id}/reviews/{userID}", s.handleDeleteReview)
	s.mux.HandleFunc("GET /api/users/{id}/reviews", s.handleListUserReviews)

	// Shelves API
	s.mux.HandleFunc("POST /api/shelves", s.handleCreateShelf)
	s.mux.HandleFunc("GET /api/shelves/{id}", s.handleGetShelfByID)
	s.mux.HandleFunc("GET /api/users/{userID}/shelves", s.handleGetUserShelves)
	s.mux.HandleFunc("PUT /api/shelves/{id}", s.handleUpdateShelf)
	s.mux.HandleFunc("DELETE /api/shelves/{id}", s.handleDeleteShelf)

	s.mux.HandleFunc("POST /api/shelves/{id}/books", s.handleAddBookToShelf)
	s.mux.HandleFunc("GET /api/shelves/{id}/books", s.handleListShelfBooks)
	s.mux.HandleFunc("DELETE /api/shelves/{id}/books/{bookID}", s.handleRemoveBookFromShelf)
	s.mux.HandleFunc("PUT /api/shelves/{id}/books/{bookID}/progress", s.handleUpdateShelfBookProgress)

	// Unknown API routes
	s.mux.HandleFunc("/api/", s.handleAPINotFound)

	// Static files
	s.registerStaticFiles()
}

func (s *Server) handleAPINotFound(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}

func (s *Server) registerStaticFiles() {
	dist, err := fs.Sub(static.Files, "dist")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(dist))
	s.mux.Handle("/", fileServer)
}
