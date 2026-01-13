-- name: CreateUser :one
INSERT INTO users (email, username, password_hash)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ?;

-- name: UpdateUser :one
UPDATE users SET email = ?, username = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?;
