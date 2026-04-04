-- name: CreateUser :one
INSERT INTO users (id, name, hashed_password, balance, created_at, updated_at)
VALUES (
  gen_random_uuid(),
  $1,
  $2,
  500,
  NOW(),
  NOW()
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

-- name: GetUserByName :one
SELECT * FROM users
WHERE name = $1;
