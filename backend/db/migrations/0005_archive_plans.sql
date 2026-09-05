-- +goose Up
ALTER TABLE training_plans ADD COLUMN archived_at TIMESTAMP;

-- +goose Down
ALTER TABLE training_plans DROP COLUMN archived_at;
