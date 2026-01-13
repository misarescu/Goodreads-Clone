-- name: CreateShelf :one
INSERT INTO shelves (user_id, name, is_default)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetShelfByID :one
SELECT * FROM shelves WHERE id = ?;

-- name: GetShelfByUserAndName :one
SELECT * FROM shelves WHERE user_id = ? AND name = ?;

-- name: ListShelvesByUser :many
SELECT * FROM shelves WHERE user_id = ? ORDER BY is_default DESC, name;

-- name: UpdateShelf :one
UPDATE shelves SET name = ?
WHERE id = ? AND is_default = FALSE
RETURNING *;

-- name: DeleteShelf :exec
DELETE FROM shelves WHERE id = ? AND is_default = FALSE;

-- name: GetDefaultShelvesByUser :many
SELECT * FROM shelves WHERE user_id = ? AND is_default = TRUE;
