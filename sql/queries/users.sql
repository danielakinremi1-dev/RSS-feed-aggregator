-- name: CreateUser :one
INSERT INTO users (name)
VALUES ($1)
RETURNING *;

-- name: GetUser :one

SELECT * FROM users
WHERE name = $1 LIMIT 1;

-- name: DeleteUsers :exec
DELETE FROM users;

-- name: GetUsers :many
SELECT name FROM users;

-- name: GetUserID :one

SELECT id FROM users
WHERE name = $1 LIMIT 1;