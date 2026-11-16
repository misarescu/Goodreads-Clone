package server

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"goodreads-clone.com/internal/business"
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
	if errors.Is(err, business.ErrDefaultShelf) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
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

	err = s.business.DeleteShelf(r.Context(), id)
	if errors.Is(err, business.ErrDefaultShelf) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		handleDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

var shelvesExportHeader = []string{
	"shelf", "title", "author", "isbn", "pages_read", "progress_percent", "finished_at", "added_at",
}

func (s *Server) handleExportShelves(w http.ResponseWriter, r *http.Request) {
	userID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	rows, err := s.business.ListShelvesForExport(r.Context(), userID)
	if err != nil {
		handleDBError(w, err)
		return
	}

	var buf bytes.Buffer
	cw := csv.NewWriter(&buf)
	_ = cw.Write(shelvesExportHeader)
	for _, row := range rows {
		finished := ""
		if row.FinishedAt != nil {
			finished = row.FinishedAt.Format(time.RFC3339)
		}
		_ = cw.Write([]string{
			row.Shelf, row.Title, row.Author, row.ISBN,
			strconv.FormatInt(row.PagesRead, 10), strconv.FormatInt(row.ProgressPercent, 10),
			finished, row.AddedAt.Format(time.RFC3339),
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		log.Printf("Error encoding CSV: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="shelves-%d.csv"`, userID))
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Printf("Error writing CSV response: %v", err)
	}
}
