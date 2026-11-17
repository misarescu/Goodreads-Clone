package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tagPage struct {
	Tags []struct {
		Tag   string `json:"tag"`
		Books int64  `json:"books"`
	} `json:"tags"`
	Total  int64 `json:"total"`
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}

func listTags(t *testing.T, s *Server, query string) tagPage {
	t.Helper()
	rec := do(t, s, http.MethodGet, "/api/tags"+query, "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.NotContains(t, rec.Body.String(), `"tags":null`)
	var out tagPage
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	return out
}

func (p tagPage) pairs() []string {
	out := make([]string, 0, len(p.Tags))
	for _, x := range p.Tags {
		out = append(out, x.Tag+":"+strconvID(x.Books))
	}
	return out
}

func setTags(t *testing.T, s *Server, id int64, tags string) {
	t.Helper()
	rec := do(t, s, http.MethodPut, "/api/books/"+strconvID(id)+"/tags", `{"tags":`+tags+`}`)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
}

func TestHandleListTagsEmpty(t *testing.T) {
	s := newTagServer(t)
	rec := do(t, s, http.MethodGet, "/api/tags", "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"tags":[],"total":0,"limit":20,"offset":0}`, rec.Body.String())
}

func TestHandleListTagsCounts(t *testing.T) {
	s := newTagServer(t)
	a := createBook(t, s, "A", "1")
	b := createBook(t, s, "B", "2")
	c := createBook(t, s, "C", "3")
	setTags(t, s, a, `["fantasy","classic","zed"]`)
	setTags(t, s, b, `["fantasy","classic"]`)
	setTags(t, s, c, `["fantasy","apple"]`)

	// books desc, then tag
	assert.Equal(t, []string{"fantasy:3", "classic:2", "apple:1", "zed:1"}, listTags(t, s, "").pairs())

	// removing from one book lowers the count; from the last removes the tag
	setTags(t, s, b, `["fantasy"]`)
	assert.Equal(t, []string{"fantasy:3", "apple:1", "classic:1", "zed:1"}, listTags(t, s, "").pairs())
	setTags(t, s, a, `["fantasy"]`)
	assert.Equal(t, []string{"fantasy:3", "apple:1"}, listTags(t, s, "").pairs())

	// deleting a book lowers the counts of its tags
	require.Equal(t, http.StatusNoContent, do(t, s, http.MethodDelete, "/api/books/"+strconvID(c), "").Code)
	assert.Equal(t, []string{"fantasy:2"}, listTags(t, s, "").pairs())
}

func TestHandleListTagsPagination(t *testing.T) {
	s := newTagServer(t)
	a := createBook(t, s, "A", "1")
	b := createBook(t, s, "B", "2")
	setTags(t, s, a, `["c","b","a"]`)
	setTags(t, s, b, `["c"]`)

	p := listTags(t, s, "?limit=2")
	assert.Equal(t, []string{"c:2", "a:1"}, p.pairs())
	assert.EqualValues(t, 3, p.Total)
	assert.EqualValues(t, 2, p.Limit)
	assert.EqualValues(t, 0, p.Offset)

	p = listTags(t, s, "?limit=2&offset=2")
	assert.Equal(t, []string{"b:1"}, p.pairs())
	assert.EqualValues(t, 3, p.Total)
	assert.EqualValues(t, 2, p.Offset)

	p = listTags(t, s, "?offset=10")
	assert.Equal(t, []string{}, p.pairs())
	assert.EqualValues(t, 3, p.Total)

	assert.EqualValues(t, 100, listTags(t, s, "?limit=500").Limit)
}

func TestHandleListTagsPrefix(t *testing.T) {
	s := newTagServer(t)
	a := createBook(t, s, "A", "1")
	b := createBook(t, s, "B", "2")
	setTags(t, s, a, `["fantasy","fan-fic","classic"]`)
	setTags(t, s, b, `["fantasy","fa"]`)

	p := listTags(t, s, "?prefix=Fan")
	assert.Equal(t, []string{"fantasy:2", "fan-fic:1"}, p.pairs())
	assert.EqualValues(t, 2, p.Total)
	assert.Equal(t, []string{"fantasy:2", "fan-fic:1"}, listTags(t, s, "?prefix=%20FAN%20").pairs())
	assert.Equal(t, []string{"fantasy:2", "classic:1", "fa:1", "fan-fic:1"}, listTags(t, s, "?prefix=").pairs())
	assert.Equal(t, []string{}, listTags(t, s, "?prefix=zzz").pairs())
	assert.EqualValues(t, 0, listTags(t, s, "?prefix=zzz").Total)
}

func TestHandleListTagsInvalidPrefix(t *testing.T) {
	s := newTagServer(t)
	for _, p := range []string{"a%20b", "a_b", "caf%C3%A9", "a%25", strings.Repeat("a", 33)} {
		rec := do(t, s, http.MethodGet, "/api/tags?prefix="+p, "")
		assert.Equal(t, http.StatusBadRequest, rec.Code, p)
		assert.Contains(t, rec.Body.String(), "Error", p)
	}
}

type similarPage struct {
	Books []struct {
		Book struct {
			ID   int64    `json:"id"`
			Tags []string `json:"tags"`
		} `json:"book"`
		SharedTags []string `json:"sharedTags"`
	} `json:"books"`
	Total  int64 `json:"total"`
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}

func listSimilar(t *testing.T, s *Server, id int64, query string) similarPage {
	t.Helper()
	rec := do(t, s, http.MethodGet, "/api/books/"+strconvID(id)+"/similar"+query, "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.NotContains(t, rec.Body.String(), `"books":null`)
	var out similarPage
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	return out
}

func TestHandleListSimilarBooks(t *testing.T) {
	s := newTagServer(t)
	a := createBook(t, s, "A", "1")
	b := createBook(t, s, "B", "2")
	c := createBook(t, s, "C", "3")
	d := createBook(t, s, "D", "4")
	setTags(t, s, a, `["fantasy","epic","zed"]`)
	setTags(t, s, b, `["fantasy"]`)
	setTags(t, s, c, `["fantasy","epic"]`)
	setTags(t, s, d, `["epic"]`)

	p := listSimilar(t, s, a, "")
	require.Len(t, p.Books, 3)
	assert.EqualValues(t, 3, p.Total)
	assert.EqualValues(t, 20, p.Limit)
	assert.Equal(t, c, p.Books[0].Book.ID)
	assert.Equal(t, []string{"epic", "fantasy"}, p.Books[0].SharedTags)
	assert.Equal(t, []string{"epic", "fantasy"}, p.Books[0].Book.Tags)
	assert.Equal(t, b, p.Books[1].Book.ID)
	assert.Equal(t, d, p.Books[2].Book.ID)

	p = listSimilar(t, s, a, "?limit=1&offset=1")
	require.Len(t, p.Books, 1)
	assert.Equal(t, b, p.Books[0].Book.ID)
	assert.EqualValues(t, 3, p.Total)
	assert.EqualValues(t, 1, p.Offset)

	setTags(t, s, c, `[]`)
	assert.EqualValues(t, 2, listSimilar(t, s, a, "").Total)
	require.Equal(t, http.StatusNoContent, do(t, s, http.MethodDelete, "/api/books/"+strconvID(b), "").Code)
	assert.EqualValues(t, 1, listSimilar(t, s, a, "").Total)
}

func TestHandleListSimilarBooksEmpty(t *testing.T) {
	s := newTagServer(t)
	a := createBook(t, s, "A", "1")
	rec := do(t, s, http.MethodGet, "/api/books/"+strconvID(a)+"/similar", "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"books":[],"total":0,"limit":20,"offset":0}`, rec.Body.String())
}

func TestHandleListSimilarBooksErrors(t *testing.T) {
	s := newTagServer(t)
	assert.Equal(t, http.StatusNotFound, do(t, s, http.MethodGet, "/api/books/999/similar", "").Code)
	assert.Equal(t, http.StatusBadRequest, do(t, s, http.MethodGet, "/api/books/abc/similar", "").Code)
}
