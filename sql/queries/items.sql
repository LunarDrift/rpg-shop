-- name: CreateItem :one
INSERT INTO items (id, name, description, price, quantity, created_at, updated_at)
VALUES (
  gen_random_uuid(),
  $1,
  $2,
  $3,
  $4,
  NOW(),
  NOW()
)
RETURNING *;

-- name: GetAllItems :many
SELECT * FROM items
ORDER BY name;

-- name: GetItemByID :one
SELECT * FROM items
WHERE id = $1;

-- name: DeleteItemByID :exec
DELETE FROM items
WHERE id = $1;

-- name: UpdateQuantity :one
UPDATE items
SET quantity = $1,
    updated_at = NOW()
WHERE id = $2
RETURNING *;
