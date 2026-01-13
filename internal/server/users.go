package server

import (
	"net/http"
	"time"

	"goodreads-clone.com/internal/db"
)

type createUserRequest struct {
	Email        string `json:"email"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

type updateUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

type userResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toUserResponse(u *db.User) userResponse {
	return userResponse{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}
}

func toUserResponses(users []db.User) []userResponse {
	response := make([]userResponse, len(users))
	for i, u := range users {
		response[i] = toUserResponse(&u)
	}
	return response
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Username == "" || req.PasswordHash == "" {
		writeError(w, http.StatusBadRequest, "email, username, and password_hash are required")
		return
	}

	user, err := s.business.CreateUser(r.Context(), req.Email, req.Username, req.PasswordHash)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toUserResponse(user))
}

func (s *Server) handleGetUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := s.business.GetUserByID(r.Context(), id)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (s *Server) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	if email := query.Get("email"); email != "" {
		user, err := s.business.GetUserByEmail(r.Context(), email)
		if err != nil {
			handleDBError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, toUserResponse(user))
		return
	}

	if username := query.Get("username"); username != "" {
		user, err := s.business.GetUserByUsername(r.Context(), username)
		if err != nil {
			handleDBError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, toUserResponse(user))
		return
	}

	limit, offset := getPaginationParams(r)
	users, err := s.business.ListUsers(r.Context(), limit, offset)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponses(users))
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req updateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := s.business.UpdateUser(r.Context(), id, req.Email, req.Username)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if err := s.business.DeleteUser(r.Context(), id); err != nil {
		handleDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
