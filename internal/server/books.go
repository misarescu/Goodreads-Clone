// Package server implements the HTTP server and request handlers for the Goodreads clone API.
package server

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"

	"goodreads-clone.com/internal/business"
	"goodreads-clone.com/internal/db"
)

type createBookRequest struct {
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	PageCount   *int64  `json:"page_count,omitempty"`
	ISBN        *string `json:"isbn,omitempty"`
	ReleaseDate *string `json:"release_date,omitempty"`
	Language    *string `json:"language,omitempty"`
	CoverURL    *string `json:"cover_url,omitempty"`
}

type updateBookRequest struct {
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	PageCount   *int64   `json:"page_count,omitempty"`
	ISBN        *string  `json:"isbn,omitempty"`
	ReleaseDate *string  `json:"release_date,omitempty"`
	Language    *string  `json:"language,omitempty"`
	CoverURL    *string  `json:"cover_url,omitempty"`
	Rating      *float64 `json:"rating,omitempty"`
}

// ratingParam validates the optional cover_url and rating and converts the rating for storage.
func (req *updateBookRequest) ratingParam() (sql.NullFloat64, error) {
	if err := validateCoverURL(req.CoverURL); err != nil {
		return sql.NullFloat64{}, err
	}
	if req.Rating == nil {
		return sql.NullFloat64{}, nil
	}
	if !validRating(*req.Rating) {
		return sql.NullFloat64{}, errors.New("rating must be between 0 and 5 in increments of 0.25")
	}
	return sql.NullFloat64{Float64: *req.Rating, Valid: true}, nil
}

// validRating reports whether r is within 0..5 in increments of 0.25.
func validRating(r float64) bool {
	return r >= 0 && r <= 5 && r*4 == math.Trunc(r*4)
}

// maxCoverURLLength is the longest accepted cover_url.
const maxCoverURLLength = 2048

// validateCoverURL checks that a cover_url is an absolute http or https URL of at
// most maxCoverURLLength characters.
func validateCoverURL(raw *string) error {
	if raw == nil {
		return nil
	}
	err := errors.New("cover_url must be an absolute http or https URL of at most 2048 characters")
	if len(*raw) > maxCoverURLLength {
		return err
	}
	u, perr := url.Parse(*raw)
	if perr != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return err
	}
	return nil
}

func (req *createBookRequest) toParams() (db.CreateBookParams, error) {
	params := db.CreateBookParams{
		Title:    req.Title,
		Author:   req.Author,
		CoverUrl: req.CoverURL,
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
			return params, fmt.Errorf("parsing release date: %w", err)
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

	if err := validateCoverURL(req.CoverURL); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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

	s.writeBookWithTags(w, r, http.StatusCreated, book)
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
	withStats, err := s.business.WithReviewStat(r.Context(), book)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, withStats)
}

// searchBooksResponse is a page of search results with the total number of matches.
type searchBooksResponse struct {
	Books  []business.BookWithTags `json:"books"`
	Total  int64                   `json:"total"`
	Limit  int64                   `json:"limit"`
	Offset int64                   `json:"offset"`
}

func (s *Server) handleSearchBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "search query 'q' is required")
		return
	}

	limit, offset := getPaginationParams(r)

	books, total, err := s.business.SearchBooks(r.Context(), query, limit, offset)
	if errors.Is(err, business.ErrInvalidSearch) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		handleDBError(w, err)
		return
	}

	withTags, err := s.business.WithTags(r.Context(), books)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, searchBooksResponse{Books: withTags, Total: total, Limit: limit, Offset: offset})
}

func (s *Server) handleGetBooks(w http.ResponseWriter, r *http.Request) {
	if isbn := r.URL.Query().Get("isbn"); isbn != "" {
		book, err := s.business.GetBookByISBN(r.Context(), isbn)
		if err != nil {
			handleDBError(w, err)
			return
		}
		withStats, err := s.business.WithReviewStat(r.Context(), book)
		if err != nil {
			handleDBError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, withStats)
		return
	}

	limit, offset := getPaginationParams(r)
	var books []db.Book
	var err error
	if r.URL.Query().Has("tag") {
		books, err = s.business.ListBooksByTag(r.Context(), r.URL.Query().Get("tag"), limit, offset)
	} else {
		books, err = s.business.ListBooks(r.Context(), limit, offset)
	}
	if errors.Is(err, business.ErrInvalidTag) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		handleDBError(w, err)
		return
	}
	withStats, err := s.business.WithReviewStats(r.Context(), books)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, withStats)
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
		ID:       id,
		Title:    req.Title,
		Author:   req.Author,
		CoverUrl: req.CoverURL,
	}

	rating, err := req.ratingParam()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	params.Rating = rating

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

	s.writeBookWithTags(w, r, http.StatusOK, book)
}

type setTagsRequest struct {
	Tags []string `json:"tags"`
}

func (s *Server) handleSetBookTags(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book ID")
		return
	}

	var req setTagsRequest
	if err := decodeJSON(r, &req); err != nil || req.Tags == nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	book, err := s.business.SetBookTags(r.Context(), id, req.Tags)
	if errors.Is(err, business.ErrInvalidTag) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
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

func (s *Server) handleCountBooks(w http.ResponseWriter, r *http.Request) {
	n, err := s.business.CountBooks(r.Context())
	if err != nil {
		handleDBError(w, err)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	// Field name book_count chosen by a human via ask_human (gc-6).
	writeJSON(w, http.StatusOK, map[string]int64{"book_count": n})
}

// writeBookWithTags responds with the book and its tags.
func (s *Server) writeBookWithTags(w http.ResponseWriter, r *http.Request, status int, book *db.Book) {
	withTags, err := s.business.WithTag(r.Context(), book)
	if err != nil {
		handleDBError(w, err)
		return
	}
	writeJSON(w, status, withTags)
}
