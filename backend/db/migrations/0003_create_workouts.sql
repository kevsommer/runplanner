-- +goose Up
CREATE TABLE IF NOT EXISTS workouts (
  id TEXT PRIMARY KEY,
  plan_id TEXT NOT NULL REFERENCES training_plans(id) ON DELETE CASCADE,
  runType TEXT NOT NULL,
  day TEXT NOT NULL,
  description TEXT NOT NULL,
  notes TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending',
  distance REAL NOT NULL
);

CREATE INDEX idx_workouts_plan_id ON workouts(plan_id);

-- +goose Down
DROP TABLE IF EXISTS workouts;

