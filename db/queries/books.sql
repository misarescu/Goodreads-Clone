-- name: CreateBook :one
INSERT INTO books (title, author, page_count, isbn, release_date, language)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetBookByID :one
SELECT * FROM books WHERE id = ?;

-- name: GetBookByISBN :one
SELECT * FROM books WHERE isbn = ?;

-- name: UpdateBook :one
UPDATE books SET title = ?, author = ?, page_count = ?, isbn = ?, release_date = ?, language = ?
WHERE id = ?
RETURNING *;

-- name: DeleteBook :exec
DELETE FROM books WHERE id = ?;

-- name: ListBooks :many
SELECT * FROM books ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: SearchBooksByTitle :many
SELECT * FROM books WHERE title LIKE ? ORDER BY title LIMIT ?;
