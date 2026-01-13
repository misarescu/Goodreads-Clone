package server

import (
	"net/http"
)

type createShelfRequest struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
}

type updateShelfRequest struct {
	Name string `json:"name"`
}

func (s *Server) handleCreateShelf(w http.ResponseWriter, r *http.Request) {
	var req createShelfRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserID == 0 || req.Name == "" {
		writeError(w, http.StatusBadRequest, "user_id and name are required")
		return
	}

	shelf, err := s.business.CreateShelf(r.Context(), req.UserID, req.Name)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, shelf)
}

func (s *Server) handleGetShelfByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid shelf ID")
		return
	}

	shelf, err := s.business.GetShelfByID(r.Context(), id)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, shelf)
}

func (s *Server) handleGetUserShelves(w http.ResponseWriter, r *http.Request) {
	userID, err := parseID(r, "userID")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if name := r.URL.Query().Get("name"); name != "" {
		shelf, err := s.business.GetShelfByUserAndName(r.Context(), userID, name)
		if err != nil {
			handleDBError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, shelf)
		return
	}

	shelves, err := s.business.ListShelvesByUser(r.Context(), userID)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, shelves)
}

func (s *Server) handleUpdateShelf(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid shelf ID")
		return
	}

	var req updateShelfRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	shelf, err := s.business.UpdateShelf(r.Context(), id, req.Name)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, shelf)
}

func (s *Server) handleDeleteShelf(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid shelf ID")
		return
	}

	if err := s.business.DeleteShelf(r.Context(), id); err != nil {
		handleDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
