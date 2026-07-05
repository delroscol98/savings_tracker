-- +goose Up
ALTER TABLE users
ADD COLUMN IF NOT EXISTS full_name TEXT NOT NULL;

-- +goose Down
ALTER TABLE users
DROP COLUMN full_name;
