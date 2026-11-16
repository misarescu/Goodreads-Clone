-- name: CreateBook :one
INSERT INTO books (title, author, page_count, isbn, release_date, language, cover_url)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetBookByID :one
SELECT * FROM books WHERE id = ?;

-- name: GetBookByISBN :one
SELECT * FROM books WHERE isbn = ?;

-- name: UpdateBook :one
UPDATE books SET title = ?, author = ?, page_count = ?, isbn = ?, release_date = ?, language = ?, cover_url = ?, rating = ?
WHERE id = ?
RETURNING *;

-- name: DeleteBook :exec
DELETE FROM books WHERE id = ?;

-- name: ListBooks :many
SELECT * FROM books ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: CountBooks :one
SELECT COUNT(*) FROM books;

-- Full-text search (SearchBooks, CountSearchBooks) is hand-written in
-- internal/storage/books.go: sqlc cannot parse FTS5's MATCH on the table name.
