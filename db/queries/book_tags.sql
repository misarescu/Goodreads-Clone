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
