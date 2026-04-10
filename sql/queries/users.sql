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

-- name: GetUserInventory :many
SELECT items.name, items.price, user_items.quantity
FROM user_items
JOIN items ON items.id = user_items.item_id
WHERE user_items.user_id = $1;

-- name: AddToInventory :exec
INSERT INTO user_items (user_id, item_id, quantity)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, item_id)
DO UPDATE SET quantity = user_items.quantity + EXCLUDED.quantity;

-- name: RemoveFromInventory :exec
DELETE FROM user_items
WHERE user_id = $1 AND item_id = $2;

-- name: UpdateInventoryQuantity :one
UPDATE user_items
SET quantity = $3
WHERE user_id = $1 AND item_id = $2
RETURNING *;
