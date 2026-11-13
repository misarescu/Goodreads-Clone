package server

import (
	"io/fs"
	"net/http"

	"goodreads-clone.com/static"
)

func (s *Server) registerRoutes() {
	// Health check
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)

	// Users API
	s.mux.HandleFunc("POST /api/users", s.handleCreateUser)
	s.mux.HandleFunc("GET /api/users/{id}", s.handleGetUserByID)
	s.mux.HandleFunc("GET /api/users", s.handleGetUsers)
	s.mux.HandleFunc("PUT /api/users/{id}", s.handleUpdateUser)
	s.mux.HandleFunc("DELETE /api/users/{id}", s.handleDeleteUser)

	// Books API
	s.mux.HandleFunc("POST /api/books", s.handleCreateBook)
	s.mux.HandleFunc("GET /api/books/search", s.handleSearchBooks)
	s.mux.HandleFunc("GET /api/books/{id}", s.handleGetBookByID)
	s.mux.HandleFunc("GET /api/books", s.handleGetBooks)
	s.mux.HandleFunc("PUT /api/books/{id}", s.handleUpdateBook)
	s.mux.HandleFunc("DELETE /api/books/{id}", s.handleDeleteBook)

	// Shelves API
	s.mux.HandleFunc("POST /api/shelves", s.handleCreateShelf)
	s.mux.HandleFunc("GET /api/shelves/{id}", s.handleGetShelfByID)
	s.mux.HandleFunc("GET /api/users/{userID}/shelves", s.handleGetUserShelves)
	s.mux.HandleFunc("PUT /api/shelves/{id}", s.handleUpdateShelf)
	s.mux.HandleFunc("DELETE /api/shelves/{id}", s.handleDeleteShelf)

	// Static files
	s.registerStaticFiles()
}

func (s *Server) registerStaticFiles() {
	dist, err := fs.Sub(static.Files, "dist")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(dist))
	s.mux.Handle("/", fileServer)
}
