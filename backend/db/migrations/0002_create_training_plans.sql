-- +goose Up
CREATE TABLE IF NOT EXISTS training_plans (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  end_date TEXT NOT NULL,
  weeks INTEGER NOT NULL,
  start_date TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_training_plans_user_id ON training_plans(user_id);

-- +goose Down
DROP TABLE IF EXISTS training_plans;
