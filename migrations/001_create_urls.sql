-- +goose Up
CREATE TABLE IF NOT EXISTS url (
    alias VARCHAR(32) PRIMARY KEY CHECK (alias ~ '^[A-Za-z0-9_-]+$'),
    original_url TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS url;
