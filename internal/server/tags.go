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
