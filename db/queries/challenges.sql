-- name: UpsertReadingChallenge :one
INSERT INTO reading_challenges (user_id, year, goal)
VALUES (?, ?, ?)
ON CONFLICT (user_id, year) DO UPDATE SET
    goal = excluded.goal,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetReadingChallenge :one
SELECT * FROM reading_challenges WHERE user_id = ? AND year = ?;

-- name: CountBooksFinishedInYear :one
SELECT COUNT(DISTINCT sb.book_id)
FROM shelf_books sb
JOIN shelves s ON s.id = sb.shelf_id
WHERE s.user_id = sqlc.arg(user_id)
  AND sb.finished_at IS NOT NULL
  AND CAST(strftime('%Y', sb.finished_at) AS INTEGER) = CAST(sqlc.arg(year) AS INTEGER);
