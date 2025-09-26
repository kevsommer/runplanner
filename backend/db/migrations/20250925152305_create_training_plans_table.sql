-- +goose Up
-- +goose StatementBegin
CREATE TABLE training_plans (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id            TEXT NOT NULL,
  goal               TEXT NOT NULL CHECK (goal IN ('5K','10K','HALF','MARATHON')),
  start_date         DATE NOT NULL,
  end_date           DATE NOT NULL CHECK (julianday(end_date) >= julianday(start_date)),
  activities_per_week INTEGER NOT NULL CHECK (activities_per_week BETWEEN 1 AND 14),
  name               TEXT,
  created_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS training_plans;
-- +goose StatementEnd
