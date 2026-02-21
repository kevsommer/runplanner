-- +goose Up
ALTER TABLE users ADD COLUMN active_plan_id TEXT REFERENCES training_plans(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE users DROP COLUMN active_plan_id;
