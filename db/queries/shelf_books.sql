-- name: AddBookToShelf :one
INSERT INTO shelf_books (shelf_id, book_id, pages_read, progress_percent)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetShelfBook :one
SELECT * FROM shelf_books WHERE shelf_id = ? AND book_id = ?;

-- name: UpdateShelfBookProgress :one
UPDATE shelf_books SET
    pages_read = sqlc.arg(pages_read),
    progress_percent = sqlc.arg(progress_percent),
    finished_at = CASE
        WHEN sqlc.arg(progress_percent) >= 100 THEN COALESCE(finished_at, CURRENT_TIMESTAMP)
        ELSE NULL
    END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: RemoveBookFromShelf :exec
DELETE FROM shelf_books WHERE shelf_id = ? AND book_id = ?;

-- name: ListBooksOnShelf :many
SELECT sb.*, b.title, b.author, b.page_count
FROM shelf_books sb
JOIN books b ON sb.book_id = b.id
WHERE sb.shelf_id = ?
ORDER BY sb.added_at DESC;

-- name: GetUserBookShelves :many
SELECT s.* FROM shelves s
JOIN shelf_books sb ON s.id = sb.shelf_id
WHERE s.user_id = ? AND sb.book_id = ?;

-- name: GetUserReadingStats :one
SELECT
    COUNT(DISTINCT sb.book_id) AS books_on_shelves,
    CAST(COALESCE(SUM(sb.pages_read), 0) AS INTEGER) AS pages_read,
    CAST(COALESCE(SUM(CASE WHEN sb.progress_percent = 100 THEN 1 ELSE 0 END), 0) AS INTEGER) AS books_finished
FROM shelf_books sb
JOIN shelves s ON s.id = sb.shelf_id
WHERE s.user_id = ?;
