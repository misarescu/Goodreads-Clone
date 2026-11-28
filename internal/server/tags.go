package server

import (
	"errors"
	"net/http"

	"goodreads-clone.com/internal/business"
	"goodreads-clone.com/internal/storage"
)

// listTagsResponse is a page of tags in use with the total number of matching tags.
type listTagsResponse struct {
	Tags   []storage.TagCount `json:"tags"`
	Total  int64              `json:"total"`
	Limit  int64              `json:"limit"`
	Offset int64              `json:"offset"`
}

func (s *Server) handleListTags(w http.ResponseWriter, r *http.Request) {
	limit, offset := getPaginationParams(r)

	tags, total, err := s.business.ListTags(r.Context(), r.URL.Query().Get("prefix"), limit, offset)
	if errors.Is(err, business.ErrInvalidTag) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, listTagsResponse{Tags: tags, Total: total, Limit: limit, Offset: offset})
}

// similarBooksResponse is a page of the books similar to a book.
type similarBooksResponse struct {
	Books  []business.SimilarBook `json:"books"`
	Total  int64                  `json:"total"`
	Limit  int64                  `json:"limit"`
	Offset int64                  `json:"offset"`
}

func (s *Server) handleListSimilarBooks(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book ID")
		return
	}
	limit, offset := getPaginationParams(r)

	books, total, err := s.business.ListSimilarBooks(r.Context(), id, limit, offset)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, similarBooksResponse{Books: books, Total: total, Limit: limit, Offset: offset})
}

func (s *Server) handleListUserTags(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}
	limit, offset := getPaginationParams(r)

	tags, total, err := s.business.ListUserTags(r.Context(), id, limit, offset)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, listTagsResponse{Tags: tags, Total: total, Limit: limit, Offset: offset})
}

// tagBooksResponse is a page of the books carrying a tag.
type tagBooksResponse struct {
	Books  []business.BookWithStats `json:"books"`
	Total  int64                    `json:"total"`
	Limit  int64                    `json:"limit"`
	Offset int64                    `json:"offset"`
}

func (s *Server) handleListTagBooks(w http.ResponseWriter, r *http.Request) {
	limit, offset := getPaginationParams(r)

	books, total, err := s.business.ListTagBooks(r.Context(), r.PathValue("tag"), limit, offset)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, tagBooksResponse{Books: books, Total: total, Limit: limit, Offset: offset})
}
