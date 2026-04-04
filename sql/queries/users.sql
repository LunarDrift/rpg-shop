-- name: CreateUser :one
INSERT INTO users (id, name, balance)
VALUES (
  gen_random_uuid(),
  $1,
  $2
  )
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetAllUsers :many
SELECT * FROM users;

-- name: UpdateBalance :one
UPDATE users
SET balance = $2
WHERE id = $1
RETURNING *;
