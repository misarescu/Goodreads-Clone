-- name: UpsertReview :one
INSERT INTO reviews (book_id, user_id, rating, body)
VALUES (?, ?, ?, ?)
ON CONFLICT (book_id, user_id) DO UPDATE SET
    rating = excluded.rating,
    body = excluded.body,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetReview :one
SELECT * FROM reviews WHERE book_id = ? AND user_id = ?;

-- name: DeleteReview :execrows
DELETE FROM reviews WHERE book_id = ? AND user_id = ?;

-- name: ListReviewsByBook :many
SELECT r.id, r.book_id, r.user_id, u.username, r.rating, r.body, r.created_at, r.updated_at
FROM reviews r
JOIN users u ON u.id = r.user_id
WHERE r.book_id = ?
ORDER BY r.created_at DESC, r.id DESC
LIMIT ? OFFSET ?;

-- name: GetReviewStats :many
SELECT book_id, COUNT(*) AS review_count, CAST(AVG(rating) AS REAL) AS average_rating
FROM reviews
WHERE book_id IN (sqlc.slice('book_ids'))
GROUP BY book_id;
