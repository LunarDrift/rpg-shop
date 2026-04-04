-- +goose Up 
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    balance INTEGER NOT NULL DEFAULT 0
);

-- +goose Down
DROP TABLE users;
