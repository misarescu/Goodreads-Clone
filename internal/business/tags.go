package business

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"goodreads-clone.com/internal/db"
	"goodreads-clone.com/internal/storage"
)

// MaxTagsPerBook is the most tags a book can have.
const MaxTagsPerBook = 10

// ErrInvalidTag is returned for a tag that is not 1 to 32 characters of a-z, 0-9 and "-",
// or for a tag list longer than MaxTagsPerBook.
var ErrInvalidTag = errors.New("invalid tag")

var tagPattern = regexp.MustCompile(`^[a-z0-9-]{1,32}$`)

// BookWithTags is a book with its sorted tags ([] when none).
type BookWithTags struct {
	db.Book

	Tags []string `json:"tags"`
}

// NormalizeTag trims and lowercases tag and checks it is a valid tag.
func NormalizeTag(tag string) (string, error) {
	t := strings.ToLower(strings.TrimSpace(tag))
	if !tagPattern.MatchString(t) {
		return "", fmt.Errorf("%w %q: must be 1 to 32 characters from a-z, 0-9 and -", ErrInvalidTag, tag)
	}
	return t, nil
}

// normalizeTags normalizes every tag, collapses duplicates and enforces the per-book limit.
// The result is sorted.
func normalizeTags(tags []string) ([]string, error) {
	seen := make(map[string]bool, len(tags))
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		t, err := NormalizeTag(tag)
		if err != nil {
			return nil, err
		}
		if !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	if len(out) > MaxTagsPerBook {
		return nil, fmt.Errorf("%w: a book has at most %d tags", ErrInvalidTag, MaxTagsPerBook)
	}
	sort.Strings(out)
	return out, nil
}

// SetBookTags replaces a book's tags and returns the book. It returns an error wrapping
// sql.ErrNoRows when the book does not exist and one wrapping ErrInvalidTag for bad tags.
func (b *Business) SetBookTags(ctx context.Context, id int64, tags []string) (*BookWithTags, error) {
	norm, err := normalizeTags(tags)
	if err != nil {
		return nil, err
	}
	book, err := b.Storage.Books.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get book for tags: %w", err)
	}
	if err := b.Storage.Books.SetTags(ctx, id, norm); err != nil {
		return nil, fmt.Errorf("failed to set book tags: %w", err)
	}
	return &BookWithTags{Book: *book, Tags: norm}, nil
}

// ListBooksByTag lists one page of the books carrying tag (normalized first).
func (b *Business) ListBooksByTag(ctx context.Context, tag string, limit, offset int64) ([]db.Book, error) {
	t, err := NormalizeTag(tag)
	if err != nil {
		return nil, err
	}
	books, err := b.Storage.Books.ListByTag(ctx, t, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list books by tag: %w", err)
	}
	return books, nil
}

// tagsByBook loads the tags of books, returning a non-nil slice for every book.
func (b *Business) tagsByBook(ctx context.Context, books []db.Book) (map[int64][]string, error) {
	ids := make([]int64, len(books))
	for i, bk := range books {
		ids[i] = bk.ID
	}
	tags, err := b.Storage.Books.TagsFor(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get book tags: %w", err)
	}
	return tags, nil
}

// WithTags attaches each book's tags.
func (b *Business) WithTags(ctx context.Context, books []db.Book) ([]BookWithTags, error) {
	tags, err := b.tagsByBook(ctx, books)
	if err != nil {
		return nil, err
	}
	out := make([]BookWithTags, len(books))
	for i, bk := range books {
		out[i] = BookWithTags{Book: bk, Tags: orEmpty(tags[bk.ID])}
	}
	return out, nil
}

// WithTag is WithTags for a single book.
func (b *Business) WithTag(ctx context.Context, book *db.Book) (*BookWithTags, error) {
	out, err := b.WithTags(ctx, []db.Book{*book})
	if err != nil {
		return nil, err
	}
	return &out[0], nil
}

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// ListTags lists one page of the tags in use with the number of books carrying each, and the
// total number of matching tags. prefix is trimmed and lowercased; an empty one matches all.
// It returns an error wrapping ErrInvalidTag when no valid tag can start with prefix.
func (b *Business) ListTags(
	ctx context.Context, prefix string, limit, offset int64,
) ([]storage.TagCount, int64, error) {
	p := strings.ToLower(strings.TrimSpace(prefix))
	if p != "" && !tagPattern.MatchString(p) {
		return nil, 0, fmt.Errorf("%w prefix %q: must be at most 32 characters from a-z, 0-9 and -", ErrInvalidTag, prefix)
	}
	tags, total, err := b.Storage.Books.ListTagCounts(ctx, p, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tags: %w", err)
	}
	if tags == nil {
		tags = []storage.TagCount{}
	}
	return tags, total, nil
}

// SimilarBook is a book sharing tags with another, with the tags they share (sorted).
type SimilarBook struct {
	Book       BookWithStats `json:"book"`
	SharedTags []string      `json:"sharedTags"`
}

// ListSimilarBooks lists one page of the books sharing the most tags with book id, and the
// total number of similar books. It returns an error wrapping sql.ErrNoRows for an unknown book.
func (b *Business) ListSimilarBooks(ctx context.Context, id, limit, offset int64) ([]SimilarBook, int64, error) {
	if _, err := b.Storage.Books.GetByID(ctx, id); err != nil {
		return nil, 0, fmt.Errorf("failed to get book for similar books: %w", err)
	}
	similar, total, err := b.Storage.Books.ListSimilar(ctx, id, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list similar books: %w", err)
	}
	if len(similar) == 0 {
		return []SimilarBook{}, total, nil
	}
	books := make([]db.Book, len(similar))
	for i, s := range similar {
		books[i] = s.Book
	}
	withStats, err := b.WithReviewStats(ctx, books)
	if err != nil {
		return nil, 0, err
	}
	out := make([]SimilarBook, len(similar))
	for i, s := range similar {
		out[i] = SimilarBook{Book: withStats[i], SharedTags: s.SharedTags}
	}
	return out, total, nil
}

// ListUserTags lists one page of the tags carried by the books on a user's shelves with the
// number of those books carrying each, and the total number of tags. It returns an error
// wrapping sql.ErrNoRows for an unknown user.
func (b *Business) ListUserTags(ctx context.Context, userID, limit, offset int64) ([]storage.TagCount, int64, error) {
	if _, err := b.Storage.Users.GetByID(ctx, userID); err != nil {
		return nil, 0, fmt.Errorf("failed to get user for tags: %w", err)
	}
	tags, total, err := b.Storage.Books.ListUserTagCounts(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list user tags: %w", err)
	}
	if tags == nil {
		tags = []storage.TagCount{}
	}
	return tags, total, nil
}
