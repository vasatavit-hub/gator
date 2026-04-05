-- name: GetUsers :many
SELECT id, created_at, updated_at, name
FROM users;