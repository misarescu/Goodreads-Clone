package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goodreads-clone.com/internal/business"
)

func newTagServer(t *testing.T) *Server {
	t.Helper()
	b, err := business.New(business.Config{DatabasePath: filepath.Join(t.TempDir(), "t.db")})
	require.NoError(t, err)
	t.Cleanup(func() { _ = b.Close() })
	s := &Server{business: b, mux: http.NewServeMux()}
	s.registerRoutes()
	return s
}

func do(t *testing.T, s *Server, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

func createBook(t *testing.T, s *Server, title, isbn string) int64 {
	t.Helper()
	rec := do(t, s, http.MethodPost, "/api/books", `{"title":"`+title+`","author":"Someone","isbn":"`+isbn+`"}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var out struct {
		ID   int64    `json:"id"`
		Tags []string `json:"tags"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, []string{}, out.Tags, "new book has empty tags")
	return out.ID
}

func tagsOf(t *testing.T, rec *httptest.ResponseRecorder) []string {
	t.Helper()
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.NotContains(t, rec.Body.String(), `"tags":null`)
	var out struct {
		Tags []string `json:"tags"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.NotNil(t, out.Tags, rec.Body.String())
	return out.Tags
}

func TestHandleSetBookTags(t *testing.T) {
	s := newTagServer(t)
	id := createBook(t, s, "Dune", "111")
	path := "/api/books/" + strconvID(id) + "/tags"

	rec := do(t, s, http.MethodPut, path, `{"tags":[" Sci-Fi ","epic","SCI-FI"]}`)
	assert.Equal(t, []string{"epic", "sci-fi"}, tagsOf(t, rec))

	// every response that returns a book carries the tags
	assert.Equal(t, []string{"epic", "sci-fi"}, tagsOf(t, do(t, s, http.MethodGet, "/api/books/"+strconvID(id), "")))
	assert.Equal(t, []string{"epic", "sci-fi"}, tagsOf(t, do(t, s, http.MethodGet, "/api/books?isbn=111", "")))
	assert.Equal(t, []string{"epic", "sci-fi"},
		tagsOf(t, do(t, s, http.MethodPut, "/api/books/"+strconvID(id), `{"title":"Dune","author":"Someone"}`)))

	var list []struct {
		Tags []string `json:"tags"`
	}
	rec = do(t, s, http.MethodGet, "/api/books", "")
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	assert.Equal(t, []string{"epic", "sci-fi"}, list[0].Tags)

	var found struct {
		Books []struct {
			Tags []string `json:"tags"`
		} `json:"books"`
	}
	rec = do(t, s, http.MethodGet, "/api/books/search?q=dune", "")
	require.Equal(t, http.StatusOK, rec.Code)
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &found))
	require.Len(t, found.Books, 1)
	assert.Equal(t, []string{"epic", "sci-fi"}, found.Books[0].Tags)

	// replace, then clear
	assert.Equal(t, []string{"classic"}, tagsOf(t, do(t, s, http.MethodPut, path, `{"tags":["classic"]}`)))
	assert.Equal(t, []string{}, tagsOf(t, do(t, s, http.MethodPut, path, `{"tags":[]}`)))
	assert.Equal(t, []string{}, tagsOf(t, do(t, s, http.MethodGet, "/api/books/"+strconvID(id), "")))
}

func TestHandleSetBookTagsErrors(t *testing.T) {
	s := newTagServer(t)
	id := createBook(t, s, "Dune", "111")
	path := "/api/books/" + strconvID(id) + "/tags"
	assert.Equal(t, []string{"keep"}, tagsOf(t, do(t, s, http.MethodPut, path, `{"tags":["keep"]}`)))

	eleven := `["a","b","c","d","e","f","g","h","i","j","k"]`
	for name, body := range map[string]string{
		"space":    `{"tags":["not ok"]}`,
		"empty":    `{"tags":[""]}`,
		"too long": `{"tags":["` + strings.Repeat("a", 33) + `"]}`,
		"symbols":  `{"tags":["a_b"]}`,
		"eleven":   `{"tags":` + eleven + `}`,
		"no tags":  `{}`,
		"bad json": `nope`,
	} {
		rec := do(t, s, http.MethodPut, path, body)
		assert.Equal(t, http.StatusBadRequest, rec.Code, name)
	}
	assert.Contains(t, do(t, s, http.MethodPut, path, `{"tags":["not ok"]}`).Body.String(), "not ok", "names the tag")

	ten := `["a","b","c","d","e","f","g","h","i","j"]`
	assert.Len(t, tagsOf(t, do(t, s, http.MethodPut, path, `{"tags":`+ten+`}`)), 10)

	assert.Equal(t, http.StatusNotFound, do(t, s, http.MethodPut, "/api/books/999/tags", `{"tags":["a"]}`).Code)
	assert.Equal(t, http.StatusBadRequest, do(t, s, http.MethodPut, "/api/books/x/tags", `{"tags":[]}`).Code)
}

func TestHandleListBooksByTag(t *testing.T) {
	s := newTagServer(t)
	ids := make([]int64, 0, 4)
	for i, title := range []string{"A", "B", "C", "D"} {
		ids = append(ids, createBook(t, s, title, strconvID(int64(i+1))))
	}
	for _, id := range ids[:3] {
		do(t, s, http.MethodPut, "/api/books/"+strconvID(id)+"/tags", `{"tags":["fantasy"]}`)
	}
	do(t, s, http.MethodPut, "/api/books/"+strconvID(ids[3])+"/tags", `{"tags":["other"]}`)

	titles := func(path string) []string {
		rec := do(t, s, http.MethodGet, path, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var books []struct {
			Title string   `json:"title"`
			Tags  []string `json:"tags"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &books))
		out := make([]string, 0, len(books))
		for _, b := range books {
			require.NotNil(t, b.Tags)
			out = append(out, b.Title)
		}
		return out
	}
	assert.Len(t, titles("/api/books?tag=fantasy"), 3)
	assert.Len(t, titles("/api/books?tag=%20FANTASY%20"), 3, "normalized")
	assert.Equal(t, []string{"D"}, titles("/api/books?tag=other"))
	assert.Empty(t, titles("/api/books?tag=nothing"))

	all := titles("/api/books?tag=fantasy")
	assert.Equal(t, all[:2], titles("/api/books?tag=fantasy&limit=2"))
	assert.Equal(t, all[2:], titles("/api/books?tag=fantasy&limit=2&offset=2"))

	for _, q := range []string{"tag=bad%20tag", "tag=", "tag=a_b"} {
		assert.Equal(t, http.StatusBadRequest, do(t, s, http.MethodGet, "/api/books?"+q, "").Code, q)
	}
}

func TestDeleteBookCascadesTags(t *testing.T) {
	s := newTagServer(t)
	id := createBook(t, s, "Dune", "111")
	do(t, s, http.MethodPut, "/api/books/"+strconvID(id)+"/tags", `{"tags":["a"]}`)
	require.Equal(t, http.StatusNoContent, do(t, s, http.MethodDelete, "/api/books/"+strconvID(id), "").Code)
	assert.Empty(t, tagsOfList(t, s, "/api/books?tag=a"))
}

func tagsOfList(t *testing.T, s *Server, path string) []json.RawMessage {
	t.Helper()
	var out []json.RawMessage
	require.NoError(t, json.Unmarshal(do(t, s, http.MethodGet, path, "").Body.Bytes(), &out))
	return out
}

func strconvID(id int64) string { return strconv.FormatInt(id, 10) }
