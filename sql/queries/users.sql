-- name: CreateUser :one
INSERT INTO users (id, name, balance)
VALUES (
  gen_random_uuid(),
  $1,
  $2
  )
RETURNING *;
