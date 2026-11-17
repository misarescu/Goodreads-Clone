-- name: DeleteBookTags :exec
DELETE FROM book_tags WHERE book_id = ?;

-- name: InsertBookTag :exec
INSERT INTO book_tags (book_id, tag) VALUES (?, ?);

-- name: ListBookTags :many
SELECT book_id, tag FROM book_tags
WHERE book_id IN (sqlc.slice('book_ids'))
ORDER BY book_id, tag;

-- name: ListBooksByTag :many
SELECT b.* FROM books b
JOIN book_tags t ON t.book_id = b.id
WHERE t.tag = ?
ORDER BY b.created_at DESC, b.id DESC LIMIT ? OFFSET ?;

-- name: ListTagCounts :many
SELECT tag, COUNT(*) AS books FROM book_tags
WHERE substr(tag, 1, length(sqlc.arg('prefix'))) = sqlc.arg('prefix')
GROUP BY tag
ORDER BY books DESC, tag
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountTags :one
SELECT COUNT(DISTINCT tag) FROM book_tags
WHERE substr(tag, 1, length(sqlc.arg('prefix'))) = sqlc.arg('prefix');

-- name: ListSimilarBooks :many
SELECT b.*, CAST(group_concat(o.tag) AS TEXT) AS shared_tags
FROM book_tags o
JOIN book_tags m ON m.tag = o.tag AND m.book_id = sqlc.arg('book_id')
JOIN books b ON b.id = o.book_id
WHERE o.book_id <> sqlc.arg('book_id')
GROUP BY b.id
ORDER BY COUNT(*) DESC, b.id
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountSimilarBooks :one
SELECT COUNT(DISTINCT o.book_id) FROM book_tags o
JOIN book_tags m ON m.tag = o.tag AND m.book_id = sqlc.arg('book_id')
WHERE o.book_id <> sqlc.arg('book_id');
