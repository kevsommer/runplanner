-- +goose Up
CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  password_hash BLOB NOT NULL,
  created_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS users;
