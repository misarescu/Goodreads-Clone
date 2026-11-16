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
    COUNT(*) AS books_on_shelves,
    CAST(COALESCE(SUM(per_book.pages_read), 0) AS INTEGER) AS pages_read,
    CAST(COALESCE(SUM(CASE WHEN per_book.progress_percent = 100 THEN 1 ELSE 0 END), 0) AS INTEGER) AS books_finished
FROM (
    SELECT
        sb.book_id,
        MAX(sb.pages_read) AS pages_read,
        MAX(sb.progress_percent) AS progress_percent
    FROM shelf_books sb
    JOIN shelves s ON s.id = sb.shelf_id
    WHERE s.user_id = ?
    GROUP BY sb.book_id
) AS per_book;

-- name: MoveShelfBook :execrows
UPDATE shelf_books SET
    shelf_id = sqlc.arg(to_shelf_id),
    updated_at = CURRENT_TIMESTAMP
WHERE shelf_id = sqlc.arg(from_shelf_id) AND book_id = sqlc.arg(book_id);

-- name: GetBookOnShelf :one
SELECT sb.*, b.title, b.author, b.page_count
FROM shelf_books sb
JOIN books b ON sb.book_id = b.id
WHERE sb.shelf_id = ? AND sb.book_id = ?;
