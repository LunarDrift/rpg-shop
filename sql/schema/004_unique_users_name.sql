-- +goose Up
ALTER TABLE users
ADD constraint users_name_unique UNIQUE (name);

-- +goose Down
ALTER TABLE users DROP constraint users_name_unique ;
