package server

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"goodreads-clone.com/internal/business"
	"goodreads-clone.com/internal/db"
	"goodreads-clone.com/internal/mocks"
	"goodreads-clone.com/internal/storage"
)

func newShelfServer(t *testing.T) (*Server, *mocks.MockShelfRepository) {
	t.Helper()
	repo := mocks.NewMockShelfRepository(t)
	s := &Server{
		business: &business.Business{Storage: &storage.Storage{Shelves: repo}},
		mux:      http.NewServeMux(),
	}
	s.registerRoutes()
	return s, repo
}

func testShelf() *db.Shelf {
	return &db.Shelf{
		ID:        7,
		UserID:    3,
		Name:      "Favorites",
		CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}
}

func shelfRequest(s *Server, method, path, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
	}
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	return rec
}

func assertShelfBody(t *testing.T, rec *httptest.ResponseRecorder, want *db.Shelf) {
	t.Helper()
	var got db.Shelf
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode body %q: %v", rec.Body.String(), err)
	}
	if got != *want {
		t.Errorf("body = %+v, want %+v", got, *want)
	}
}

func assertErrorStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d (body %q)", rec.Code, want, rec.Body.String())
	}
}

func TestHandleCreateShelf(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, repo := newShelfServer(t)
		want := testShelf()
		repo.EXPECT().Create(mock.Anything, int64(3), "Favorites", false).Return(want, nil)

		rec := shelfRequest(s, http.MethodPost, "/api/shelves", `{"user_id":3,"name":"Favorites"}`)
		assertErrorStatus(t, rec, http.StatusCreated)
		assertShelfBody(t, rec, want)
	})

	for name, body := range map[string]string{
		"malformed JSON": `{`,
		"unknown field":  `{"user_id":3,"name":"x","bogus":1}`,
		"empty body":     ``,
		"missing name":   `{"user_id":3}`,
		"missing user":   `{"name":"x"}`,
	} {
		t.Run("400 "+name, func(t *testing.T) {
			s, _ := newShelfServer(t)
			assertErrorStatus(t, shelfRequest(s, http.MethodPost, "/api/shelves", body), http.StatusBadRequest)
		})
	}

	t.Run("500 on business error", func(t *testing.T) {
		s, repo := newShelfServer(t)
		repo.EXPECT().Create(mock.Anything, int64(3), "Favorites", false).Return(nil, errors.New("boom"))

		rec := shelfRequest(s, http.MethodPost, "/api/shelves", `{"user_id":3,"name":"Favorites"}`)
		assertErrorStatus(t, rec, http.StatusInternalServerError)
	})
}

func TestHandleGetShelfByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, repo := newShelfServer(t)
		want := testShelf()
		repo.EXPECT().GetByID(mock.Anything, int64(7)).Return(want, nil)

		rec := shelfRequest(s, http.MethodGet, "/api/shelves/7", "")
		assertErrorStatus(t, rec, http.StatusOK)
		assertShelfBody(t, rec, want)
	})

	t.Run("400 invalid ID", func(t *testing.T) {
		s, _ := newShelfServer(t)
		assertErrorStatus(t, shelfRequest(s, http.MethodGet, "/api/shelves/abc", ""), http.StatusBadRequest)
	})

	t.Run("404 not found", func(t *testing.T) {
		s, repo := newShelfServer(t)
		repo.EXPECT().GetByID(mock.Anything, int64(7)).Return(nil, sql.ErrNoRows)
		assertErrorStatus(t, shelfRequest(s, http.MethodGet, "/api/shelves/7", ""), http.StatusNotFound)
	})

	t.Run("500 on business error", func(t *testing.T) {
		s, repo := newShelfServer(t)
		repo.EXPECT().GetByID(mock.Anything, int64(7)).Return(nil, errors.New("boom"))
		assertErrorStatus(t, shelfRequest(s, http.MethodGet, "/api/shelves/7", ""), http.StatusInternalServerError)
	})
}

func TestHandleGetShelfByName(t *testing.T) {
	const path = "/api/users/3/shelves?name=Favorites"

	t.Run("success", func(t *testing.T) {
		s, repo := newShelfServer(t)
		want := testShelf()
		repo.EXPECT().GetByUserAndName(mock.Anything, int64(3), "Favorites").Return(want, nil)

		rec := shelfRequest(s, http.MethodGet, path, "")
		assertErrorStatus(t, rec, http.StatusOK)
		assertShelfBody(t, rec, want)
	})

	t.Run("400 invalid user ID", func(t *testing.T) {
		s, _ := newShelfServer(t)
		assertErrorStatus(t, shelfRequest(s, http.MethodGet, "/api/users/abc/shelves?name=x", ""), http.StatusBadRequest)
	})

	t.Run("404 not found", func(t *testing.T) {
		s, repo := newShelfServer(t)
		repo.EXPECT().GetByUserAndName(mock.Anything, int64(3), "Favorites").Return(nil, sql.ErrNoRows)
		assertErrorStatus(t, shelfRequest(s, http.MethodGet, path, ""), http.StatusNotFound)
	})

	t.Run("500 on business error", func(t *testing.T) {
		s, repo := newShelfServer(t)
		repo.EXPECT().GetByUserAndName(mock.Anything, int64(3), "Favorites").Return(nil, errors.New("boom"))
		assertErrorStatus(t, shelfRequest(s, http.MethodGet, path, ""), http.StatusInternalServerError)
	})
}

func TestHandleGetUserShelves(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, repo := newShelfServer(t)
		want := []db.Shelf{*testShelf(), {ID: 8, UserID: 3, Name: "Read", IsDefault: true, CreatedAt: testShelf().CreatedAt}}
		repo.EXPECT().ListByUser(mock.Anything, int64(3)).Return(want, nil)

		rec := shelfRequest(s, http.MethodGet, "/api/users/3/shelves", "")
		assertErrorStatus(t, rec, http.StatusOK)
		var got []db.Shelf
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
			t.Errorf("body = %+v, want %+v", got, want)
		}
	})

	t.Run("400 invalid user ID", func(t *testing.T) {
		s, _ := newShelfServer(t)
		assertErrorStatus(t, shelfRequest(s, http.MethodGet, "/api/users/abc/shelves", ""), http.StatusBadRequest)
	})

	t.Run("404 not found", func(t *testing.T) {
		s, repo := newShelfServer(t)
		repo.EXPECT().ListByUser(mock.Anything, int64(3)).Return(nil, sql.ErrNoRows)
		assertErrorStatus(t, shelfRequest(s, http.MethodGet, "/api/users/3/shelves", ""), http.StatusNotFound)
	})

	t.Run("500 on business error", func(t *testing.T) {
		s, repo := newShelfServer(t)
		repo.EXPECT().ListByUser(mock.Anything, int64(3)).Return(nil, errors.New("boom"))
		assertErrorStatus(t, shelfRequest(s, http.MethodGet, "/api/users/3/shelves", ""), http.StatusInternalServerError)
	})
}

func TestHandleUpdateShelf(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, repo := newShelfServer(t)
		want := testShelf()
		want.Name = "Renamed"
		repo.EXPECT().GetByID(mock.Anything, int64(7)).Return(testShelf(), nil)
		repo.EXPECT().GetByID(mock.Anything, int64(7)).Return(testShelf(), nil)
		repo.EXPECT().Update(mock.Anything, int64(7), "Renamed").Return(want, nil)

		rec := shelfRequest(s, http.MethodPut, "/api/shelves/7", `{"name":"Renamed"}`)
		assertErrorStatus(t, rec, http.StatusOK)
		assertShelfBody(t, rec, want)
	})

	t.Run("400 invalid ID", func(t *testing.T) {
		s, _ := newShelfServer(t)
		assertErrorStatus(t, shelfRequest(s, http.MethodPut, "/api/shelves/abc", `{"name":"x"}`), http.StatusBadRequest)
	})

	for name, body := range map[string]string{
		"malformed JSON": `{`,
		"unknown field":  `{"name":"x","bogus":1}`,
		"empty body":     ``,
	} {
		t.Run("400 "+name, func(t *testing.T) {
			s, _ := newShelfServer(t)
			assertErrorStatus(t, shelfRequest(s, http.MethodPut, "/api/shelves/7", body), http.StatusBadRequest)
		})
	}

	t.Run("404 not found", func(t *testing.T) {
		s, repo := newShelfServer(t)
		repo.EXPECT().GetByID(mock.Anything, int64(7)).Return(testShelf(), nil)
		repo.EXPECT().GetByID(mock.Anything, int64(7)).Return(testShelf(), nil)
		repo.EXPECT().Update(mock.Anything, int64(7), "x").Return(nil, sql.ErrNoRows)
		assertErrorStatus(t, shelfRequest(s, http.MethodPut, "/api/shelves/7", `{"name":"x"}`), http.StatusNotFound)
	})

	t.Run("500 on business error", func(t *testing.T) {
		s, repo := newShelfServer(t)
		repo.EXPECT().GetByID(mock.Anything, int64(7)).Return(testShelf(), nil)
		repo.EXPECT().GetByID(mock.Anything, int64(7)).Return(testShelf(), nil)
		repo.EXPECT().Update(mock.Anything, int64(7), "x").Return(nil, errors.New("boom"))
		rec := shelfRequest(s, http.MethodPut, "/api/shelves/7", `{"name":"x"}`)
		assertErrorStatus(t, rec, http.StatusInternalServerError)
	})
}

func TestHandleDeleteShelf(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, repo := newShelfServer(t)
		repo.EXPECT().GetByID(mock.Anything, int64(7)).Return(testShelf(), nil)
		repo.EXPECT().GetByID(mock.Anything, int64(7)).Return(testShelf(), nil)
		repo.EXPECT().Delete(mock.Anything, int64(7)).Return(nil)

		rec := shelfRequest(s, http.MethodDelete, "/api/shelves/7", "")
		assertErrorStatus(t, rec, http.StatusNoContent)
		if rec.Body.Len() != 0 {
			t.Errorf("body = %q, want empty", rec.Body.String())
		}
	})

	t.Run("400 invalid ID", func(t *testing.T) {
		s, _ := newShelfServer(t)
		assertErrorStatus(t, shelfRequest(s, http.MethodDelete, "/api/shelves/abc", ""), http.StatusBadRequest)
	})

	t.Run("404 not found", func(t *testing.T) {
		s, repo := newShelfServer(t)
		repo.EXPECT().GetByID(mock.Anything, int64(7)).Return(testShelf(), nil)
		repo.EXPECT().GetByID(mock.Anything, int64(7)).Return(testShelf(), nil)
		repo.EXPECT().Delete(mock.Anything, int64(7)).Return(sql.ErrNoRows)
		assertErrorStatus(t, shelfRequest(s, http.MethodDelete, "/api/shelves/7", ""), http.StatusNotFound)
	})

	t.Run("500 on business error", func(t *testing.T) {
		s, repo := newShelfServer(t)
		repo.EXPECT().GetByID(mock.Anything, int64(7)).Return(testShelf(), nil)
		repo.EXPECT().GetByID(mock.Anything, int64(7)).Return(testShelf(), nil)
		repo.EXPECT().Delete(mock.Anything, int64(7)).Return(errors.New("boom"))
		assertErrorStatus(t, shelfRequest(s, http.MethodDelete, "/api/shelves/7", ""), http.StatusInternalServerError)
	})
}

func TestHandleDefaultShelfProtected(t *testing.T) {
	def := testShelf()
	def.IsDefault = true
	for method, body := range map[string]string{http.MethodPut: `{"name":"x"}`, http.MethodDelete: ""} {
		t.Run("409 "+method, func(t *testing.T) {
			s, repo := newShelfServer(t)
			repo.EXPECT().GetByID(mock.Anything, int64(7)).Return(def, nil)
			rec := shelfRequest(s, method, "/api/shelves/7", body)
			assertErrorStatus(t, rec, http.StatusConflict)
			if !strings.Contains(rec.Body.String(), "default shelves cannot be renamed or deleted") {
				t.Errorf("body = %q", rec.Body.String())
			}
		})
	}
}

func TestExportShelves(t *testing.T) {
	s := newShelfBooksServer(t)
	for _, q := range []string{
		`INSERT INTO shelves (user_id, name) VALUES (1, 'Alpha')`,
		`INSERT INTO books (title, author, isbn) VALUES ('Dune, "Part" One', 'Herbert', '123')`,
		`INSERT INTO books (title, author) VALUES ('Aardvark', 'Zed')`,
		`INSERT INTO shelf_books (shelf_id, book_id, pages_read, progress_percent, finished_at, added_at)
		 VALUES (1, 1, 412, 100, '2024-03-01 10:00:00', '2024-01-01 00:00:00')`,
		`INSERT INTO shelf_books (shelf_id, book_id, added_at) VALUES (2, 3, '2024-02-02 00:00:00')`,
		`INSERT INTO shelf_books (shelf_id, book_id, added_at) VALUES (2, 2, '2024-02-03 00:00:00')`,
	} {
		if _, err := s.business.Storage.DB().Exec(q); err != nil {
			t.Fatalf("%q: %v", q, err)
		}
	}

	rec := doRequest(s, http.MethodGet, "/api/users/1/shelves/export.csv", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/csv; charset=utf-8" {
		t.Errorf("Content-Type = %q", got)
	}
	if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="shelves-1.csv"` {
		t.Errorf("Content-Disposition = %q", got)
	}
	want := "shelf,title,author,isbn,pages_read,progress_percent,finished_at,added_at\n" +
		"Alpha,Aardvark,Zed,,0,0,,2024-02-02T00:00:00Z\n" +
		"Alpha,\"Dune, \"\"Part\"\" One\",Herbert,123,0,0,,2024-02-03T00:00:00Z\n" +
		"Shelf,Title,Author,,412,100,2024-03-01T10:00:00Z,2024-01-01T00:00:00Z\n"
	if got := rec.Body.String(); got != want {
		t.Errorf("body =\n%s\nwant\n%s", got, want)
	}

	records, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil || len(records) != 4 || records[2][1] != `Dune, "Part" One` {
		t.Errorf("round trip: %v %v", records, err)
	}
}

func TestExportShelvesQuotesNewline(t *testing.T) {
	s := newShelfBooksServer(t)
	if _, err := s.business.Storage.DB().Exec(`UPDATE books SET title = 'Two' || char(10) || 'Lines'`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.business.Storage.DB().Exec(
		`INSERT INTO shelf_books (shelf_id, book_id, added_at) VALUES (1, 1, '2024-01-01 00:00:00')`); err != nil {
		t.Fatal(err)
	}
	rec := doRequest(s, http.MethodGet, "/api/users/1/shelves/export.csv", "")
	if !strings.Contains(rec.Body.String(), "\"Two\nLines\"") {
		t.Errorf("body = %q", rec.Body)
	}
}

func TestExportShelvesNoBooks(t *testing.T) {
	s := newShelfBooksServer(t)
	rec := doRequest(s, http.MethodGet, "/api/users/1/shelves/export.csv", "")
	want := "shelf,title,author,isbn,pages_read,progress_percent,finished_at,added_at\n"
	if rec.Code != http.StatusOK || rec.Body.String() != want {
		t.Errorf("status = %d body = %q", rec.Code, rec.Body)
	}
}

func TestExportShelvesErrors(t *testing.T) {
	s := newShelfBooksServer(t)
	if rec := doRequest(s, http.MethodGet, "/api/users/99/shelves/export.csv", ""); rec.Code != http.StatusNotFound {
		t.Errorf("unknown user status = %d", rec.Code)
	}
	if rec := doRequest(s, http.MethodGet, "/api/users/x/shelves/export.csv", ""); rec.Code != http.StatusBadRequest {
		t.Errorf("bad id status = %d", rec.Code)
	}
}
