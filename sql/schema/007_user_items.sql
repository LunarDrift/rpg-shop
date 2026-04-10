-- +goose Up
CREATE TABLE user_items (
    user_id UUID REFERENCES users (id) ON DELETE CASCADE,
    item_id UUID REFERENCES items (id) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 1,
    PRIMARY KEY (user_id, item_id)
);

-- +goose Down
DROP TABLE user_items;
