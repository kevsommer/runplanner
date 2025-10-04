-- +goose Up
-- +goose StatementBegin
CREATE TABLE planned_sessions (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  plan_id            INTEGER NOT NULL,
  date               DATE NOT NULL,
  workout_type       TEXT NOT NULL,
  target_distance_km REAL,
  status             TEXT NOT NULL DEFAULT 'scheduled',
  name	             TEXT,
  description        TEXT,
  week_index         INTEGER NOT NULL,
  FOREIGN KEY (plan_id) REFERENCES training_plans(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS planned_sessions;
-- +goose StatementEnd
