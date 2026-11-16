package server

import (
	"errors"
	"net/http"

	"goodreads-clone.com/internal/business"
	"goodreads-clone.com/internal/db"
	"modernc.org/sqlite"
	sqlitelib "modernc.org/sqlite/lib"
)

type addShelfBookRequest struct {
	BookID int64 `json:"book_id"`
}

// isUniqueViolation reports whether err is a SQLite unique constraint failure.
func isUniqueViolation(err error) bool {
	var sqlErr *sqlite.Error
	return errors.As(err, &sqlErr) &&
		(sqlErr.Code() == sqlitelib.SQLITE_CONSTRAINT_UNIQUE || sqlErr.Code() == sqlitelib.SQLITE_CONSTRAINT_PRIMARYKEY)
}

// isForeignKeyViolation reports whether err is a SQLite foreign key failure.
func isForeignKeyViolation(err error) bool {
	var sqlErr *sqlite.Error
	return errors.As(err, &sqlErr) && sqlErr.Code() == sqlitelib.SQLITE_CONSTRAINT_FOREIGNKEY
}

func (s *Server) handleAddBookToShelf(w http.ResponseWriter, r *http.Request) {
	shelfID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid shelf ID")
		return
	}

	var req addShelfBookRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.BookID <= 0 {
		writeError(w, http.StatusBadRequest, "book_id is required")
		return
	}

	if _, err := s.business.GetShelfByID(r.Context(), shelfID); err != nil {
		handleDBError(w, err)
		return
	}
	if _, err := s.business.GetBookByID(r.Context(), req.BookID); err != nil {
		handleDBError(w, err)
		return
	}

	shelfBook, err := s.business.AddBookToShelf(r.Context(), shelfID, req.BookID, 0, 0)
	if err != nil {
		switch {
		case isUniqueViolation(err):
			writeError(w, http.StatusConflict, "book is already on the shelf")
		case isForeignKeyViolation(err):
			writeError(w, http.StatusNotFound, "resource not found")
		default:
			handleDBError(w, err)
		}
		return
	}

	writeJSON(w, http.StatusCreated, shelfBook)
}

func (s *Server) handleRemoveBookFromShelf(w http.ResponseWriter, r *http.Request) {
	shelfID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid shelf ID")
		return
	}
	bookID, err := parseID(r, "bookID")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book ID")
		return
	}

	if _, err := s.business.GetShelfByID(r.Context(), shelfID); err != nil {
		handleDBError(w, err)
		return
	}
	if _, err := s.business.GetShelfBook(r.Context(), shelfID, bookID); err != nil {
		handleDBError(w, err)
		return
	}

	if err := s.business.RemoveBookFromShelf(r.Context(), shelfID, bookID); err != nil {
		handleDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListShelfBooks(w http.ResponseWriter, r *http.Request) {
	shelfID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid shelf ID")
		return
	}

	if _, err := s.business.GetShelfByID(r.Context(), shelfID); err != nil {
		handleDBError(w, err)
		return
	}

	books, err := s.business.ListBooksOnShelf(r.Context(), shelfID)
	if err != nil {
		handleDBError(w, err)
		return
	}
	if books == nil {
		books = []db.ListBooksOnShelfRow{}
	}

	writeJSON(w, http.StatusOK, books)
}

type updateProgressRequest struct {
	PagesRead       *int64 `json:"pages_read"`
	ProgressPercent *int64 `json:"progress_percent"`
}

func (s *Server) handleUpdateShelfBookProgress(w http.ResponseWriter, r *http.Request) {
	shelfID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid shelf ID")
		return
	}
	bookID, err := parseID(r, "bookID")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book ID")
		return
	}

	var req updateProgressRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	shelfBook, err := s.business.RecordReadingProgress(r.Context(), shelfID, bookID, req.PagesRead, req.ProgressPercent)
	if err != nil {
		if errors.Is(err, business.ErrInvalidProgress) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, shelfBook)
}

type moveShelfBookRequest struct {
	ToShelfID int64 `json:"to_shelf_id"`
}

func (s *Server) handleMoveShelfBook(w http.ResponseWriter, r *http.Request) {
	shelfID, err := parseID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid shelf ID")
		return
	}
	bookID, err := parseID(r, "bookID")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book ID")
		return
	}
	var req moveShelfBookRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ToShelfID <= 0 {
		writeError(w, http.StatusBadRequest, "to_shelf_id is required")
		return
	}

	moved, err := s.business.MoveBookToShelf(r.Context(), shelfID, bookID, req.ToShelfID)
	if err != nil {
		switch {
		case errors.Is(err, business.ErrSameShelf), errors.Is(err, business.ErrDifferentUsers):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, business.ErrBookAlreadyOnShelf), isUniqueViolation(err):
			writeError(w, http.StatusConflict, business.ErrBookAlreadyOnShelf.Error())
		default:
			handleDBError(w, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, moved)
}
