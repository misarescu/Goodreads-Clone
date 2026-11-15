package business

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"goodreads-clone.com/internal/db"
	"goodreads-clone.com/internal/storage"
)

// ErrInvalidSearch is returned when a search query has nothing searchable in it.
var ErrInvalidSearch = errors.New("search query needs at least one word or author: filter")

// SearchQuery is a parsed user search string.
type SearchQuery struct {
	// Match is the FTS5 MATCH expression built from the free words and the author filter.
	Match string
	// Language is the lang: filter, empty when absent.
	Language string
}

// ParseSearchQuery turns user input such as `hobbit author:tolkien lang:en` into an FTS5
// MATCH expression. Free words match title or author; the last free word is a prefix match.
// Every term is reduced to letters and digits (what FTS5's tokenizer indexes) and quoted, so
// punctuation and FTS syntax (AND, NEAR, quotes, parentheses, column filters) can never
// reach FTS5 as operators. Input with no searchable term yields ErrInvalidSearch.
func ParseSearchQuery(q string) (SearchQuery, error) {
	var out SearchQuery
	var words, authors []string
	for _, field := range strings.Fields(q) {
		lower := strings.ToLower(field)
		switch {
		case strings.HasPrefix(lower, "author:"):
			authors = append(authors, searchTerms(field[len("author:"):])...)
		case strings.HasPrefix(lower, "lang:"):
			if v := strings.TrimSpace(field[len("lang:"):]); v != "" {
				out.Language = v
			}
		default:
			words = append(words, searchTerms(field)...)
		}
	}

	out.Match = buildMatch(words, authors)
	if out.Match == "" {
		return SearchQuery{}, ErrInvalidSearch
	}
	return out, nil
}

// buildMatch joins the free words (the last one a prefix) and the author terms into a MATCH expression.
func buildMatch(words, authors []string) string {
	var parts []string
	for i, w := range words {
		term := `"` + w + `"`
		if i == len(words)-1 {
			term += "*"
		}
		parts = append(parts, term)
	}
	if len(authors) > 0 {
		// One column filter per term: FTS5 accepts only a single phrase after the colon here.
		for _, a := range authors {
			parts = append(parts, `author : "`+a+`"`)
		}
	}
	return strings.Join(parts, " ")
}

// searchTerms splits s into lowercase runs of letters and digits.
func searchTerms(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

// SearchBooks runs a full-text search and returns one page of results and the total match count.
func (b *Business) SearchBooks(ctx context.Context, q string, limit, offset int64) ([]db.Book, int64, error) {
	parsed, err := ParseSearchQuery(q)
	if err != nil {
		return nil, 0, err
	}
	books, total, err := b.Storage.Books.Search(ctx, storage.SearchParams{
		Match:    parsed.Match,
		Language: parsed.Language,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search books: %w", err)
	}
	return books, total, nil
}
