package server

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"goodreads-clone.com/internal/db"
)

type createBookRequest struct {
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	PageCount   *int64  `json:"page_count,omitempty"`
	ISBN        *string `json:"isbn,omitempty"`
	ReleaseDate *string `json:"release_date,omitempty"`
	Language    *string `json:"language,omitempty"`
}

type updateBookRequest struct {
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	PageCount   *int64  `json:"page_count,omitempty"`
	ISBN        *string `json:"isbn,omitempty"`
	ReleaseDate *string `json:"release_date,omitempty"`
	Language    *string `json:"language,omitempty"`
}

func (req *createBookRequest) toParams() (db.CreateBookParams, error) {
	params := db.CreateBookParams{
		Title:  req.Title,
		Author: req.Author,
	}

	if req.PageCount != nil {
		params.PageCount = sql.NullInt64{Int64: *req.PageCount, Valid: true}
	}
	if req.ISBN != nil {
		params.Isbn = sql.NullString{String: *req.ISBN, Valid: true}
	}
	if req.Language != nil {
		params.Language = sql.NullString{String: *req.Language, Valid: true}
	}
	if req.ReleaseDate != nil {
		t, err := time.Parse(time.RFC3339, *req.ReleaseDate)
		if err != nil {
			return params, err
		}
		params.ReleaseDate = sql.NullTime{Time: t, Valid: true}
	}

	return params, nil
}

func (s *Server) handleCreateBook(w http.ResponseWriter, r *http.Request) {
	var req createBookRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" || req.Author == "" {
		writeError(w, http.StatusBadRequest, "title and author are required")
		return
	}

	params, err := req.toParams()
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid release_date format, use RFC3339")
		return
	}

	book, err := s.business.CreateBook(r.Context(), params)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, book)
}

func (s *Server) handleGetBookByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book ID")
		return
	}

	book, err := s.business.GetBookByID(r.Context(), id)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, book)
}

func (s *Server) handleSearchBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "search query 'q' is required")
		return
	}

	limit := int64(20)
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	books, err := s.business.SearchBooksByTitle(r.Context(), query, limit)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, books)
}

func (s *Server) handleGetBooks(w http.ResponseWriter, r *http.Request) {
	if isbn := r.URL.Query().Get("isbn"); isbn != "" {
		book, err := s.business.GetBookByISBN(r.Context(), isbn)
		if err != nil {
			handleDBError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, book)
		return
	}

	limit, offset := getPaginationParams(r)
	books, err := s.business.ListBooks(r.Context(), limit, offset)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, books)
}

func (s *Server) handleUpdateBook(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book ID")
		return
	}

	var req updateBookRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	params := db.UpdateBookParams{
		ID:     id,
		Title:  req.Title,
		Author: req.Author,
	}

	if req.PageCount != nil {
		params.PageCount = sql.NullInt64{Int64: *req.PageCount, Valid: true}
	}
	if req.ISBN != nil {
		params.Isbn = sql.NullString{String: *req.ISBN, Valid: true}
	}
	if req.Language != nil {
		params.Language = sql.NullString{String: *req.Language, Valid: true}
	}
	if req.ReleaseDate != nil {
		t, err := time.Parse(time.RFC3339, *req.ReleaseDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid release_date format")
			return
		}
		params.ReleaseDate = sql.NullTime{Time: t, Valid: true}
	}

	book, err := s.business.UpdateBook(r.Context(), params)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, book)
}

func (s *Server) handleDeleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book ID")
		return
	}

	if err := s.business.DeleteBook(r.Context(), id); err != nil {
		handleDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
