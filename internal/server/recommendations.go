package server

import (
	"net/http"
	"strconv"

	"goodreads-clone.com/internal/business"
)

func (s *Server) handleGetRecommendations(w http.ResponseWriter, r *http.Request) {
	userID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	limit := int64(business.DefaultRecommendationLimit)
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, err = strconv.ParseInt(l, 10, 64)
		if err != nil || limit < 1 {
			writeError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = min(limit, business.MaxRecommendationLimit)
	}

	recs, err := s.business.GetRecommendations(r.Context(), userID, limit)
	if err != nil {
		handleDBError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, recs)
}
