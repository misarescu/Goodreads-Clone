package server

import (
	"errors"
	"net/http"

	"goodreads-clone.com/internal/business"
)

type challengeRequest struct {
	Goal *int64 `json:"goal"`
}

type challengeResponse struct {
	Year int64 `json:"year"`
	Goal int64 `json:"goal"`
}

func parseChallengePath(w http.ResponseWriter, r *http.Request) (userID, year int64, ok bool) {
	userID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return 0, 0, false
	}
	year, err = parseID(r, "year")
	if err != nil || year < 1 || year > 9999 {
		writeError(w, http.StatusBadRequest, "invalid year")
		return 0, 0, false
	}
	return userID, year, true
}

func (s *Server) handleSetChallenge(w http.ResponseWriter, r *http.Request) {
	userID, year, ok := parseChallengePath(w, r)
	if !ok {
		return
	}

	var req challengeRequest
	if err := decodeJSON(r, &req); err != nil || req.Goal == nil {
		writeError(w, http.StatusBadRequest, "goal is required")
		return
	}

	c, err := s.business.SetReadingChallenge(r.Context(), userID, year, *req.Goal)
	if err != nil {
		if errors.Is(err, business.ErrInvalidGoal) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, challengeResponse{Year: c.Year, Goal: c.Goal})
}

func (s *Server) handleGetChallenge(w http.ResponseWriter, r *http.Request) {
	userID, year, ok := parseChallengePath(w, r)
	if !ok {
		return
	}

	p, err := s.business.GetReadingChallengeProgress(r.Context(), userID, year)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, p)
}
