-- +goose Up
ALTER TABLE feeds
ADD last_updated_at TIMESTAMP;

-- +goose Down
ALTER TABLE feeds
DROP COLUMN last_updated_at;

