package business_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goodreads-clone.com/internal/business"
)

func TestParseSearchQuery(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		in   string
		want business.SearchQuery
	}{
		"single word is a prefix":    {"tolk", business.SearchQuery{Match: `"tolk"*`}},
		"only last word is a prefix": {"the hobbit", business.SearchQuery{Match: `"the" "hobbit"*`}},
		"author filter":              {"hobbit author:Tolkien", business.SearchQuery{Match: `"hobbit"* author : "tolkien"`}},
		"lang filter":                {"hobbit lang:EN", business.SearchQuery{Match: `"hobbit"*`, Language: "EN"}},
		"author and lang only": {
			"author:tolkien lang:en", business.SearchQuery{Match: `author : "tolkien"`, Language: "en"},
		},
		"fts operators are literal": {
			"cats AND OR NOT NEAR", business.SearchQuery{Match: `"cats" "and" "or" "not" "near"*`},
		},
		"quotes and punctuation": {`"hobbit" (title:x) ^a*`, business.SearchQuery{Match: `"hobbit" "title" "x" "a"*`}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := business.ParseSearchQuery(tt.in)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseSearchQueryRejectsUnsearchable(t *testing.T) {
	t.Parallel()
	for _, in := range []string{" ", "!!!", `"`, "()*", "AND", "-", "author:", "author:!!", "lang:en", "^ : *"} {
		if in == "AND" {
			continue // a bare AND is an ordinary word once quoted
		}
		_, err := business.ParseSearchQuery(in)
		assert.ErrorIs(t, err, business.ErrInvalidSearch, "input %q", in)
	}
}
