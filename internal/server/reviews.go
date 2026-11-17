package server

import (
	"errors"
	"net/http"
	"strconv"

	"goodreads-clone.com/internal/business"
)

type reviewRequest struct {
	UserID *int64  `json:"user_id"`
	Rating *int64  `json:"rating"`
	Body   *string `json:"body"`
}

func (s *Server) handleSaveReview(w http.ResponseWriter, r *http.Request) {
	bookID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book ID")
		return
	}

	var req reviewRequest
	if err := decodeJSON(r, &req); err != nil || req.UserID == nil || req.Rating == nil || req.Body == nil {
		writeError(w, http.StatusBadRequest, "user_id, rating and body are required")
		return
	}

	rev, created, err := s.business.SaveReview(r.Context(), bookID, *req.UserID, *req.Rating, *req.Body)
	if err != nil {
		if errors.Is(err, business.ErrInvalidRating) || errors.Is(err, business.ErrInvalidBody) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		handleDBError(w, err)
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, rev)
}

func (s *Server) handleListReviews(w http.ResponseWriter, r *http.Request) {
	bookID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book ID")
		return
	}

	limit, offset := getPaginationParams(r)
	sort := business.SortRecent
	if v := r.URL.Query().Get("sort"); v != "" {
		sort = v
	}
	var minRating int64
	if v := r.URL.Query().Get("min_rating"); v != "" {
		minRating, err = strconv.ParseInt(v, 10, 64)
		if err != nil || minRating < business.MinRating || minRating > business.MaxRating {
			writeError(w, http.StatusBadRequest, business.ErrInvalidMinRating.Error())
			return
		}
	}
	reviews, err := s.business.ListReviews(r.Context(), bookID, minRating, sort, limit, offset)
	if err != nil {
		if errors.Is(err, business.ErrInvalidSort) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, reviews)
}

func (s *Server) handleDeleteReview(w http.ResponseWriter, r *http.Request) {
	bookID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book ID")
		return
	}
	userID, err := parseID(r, "userID")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if err := s.business.DeleteReview(r.Context(), bookID, userID); err != nil {
		handleDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListUserReviews(w http.ResponseWriter, r *http.Request) {
	userID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	limit, offset := getPaginationParams(r)
	reviews, err := s.business.ListUserReviews(r.Context(), userID, limit, offset)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, reviews)
}

func (s *Server) handleRatingSpread(w http.ResponseWriter, r *http.Request) {
	bookID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book ID")
		return
	}
	spread, err := s.business.RatingSpread(r.Context(), bookID)
	if err != nil {
		handleDBError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, spread)
}
