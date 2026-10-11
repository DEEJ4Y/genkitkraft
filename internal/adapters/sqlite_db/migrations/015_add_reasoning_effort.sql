-- +goose Up
ALTER TABLE agents ADD COLUMN reasoning_effort_enabled BOOLEAN NOT NULL DEFAULT 0;
ALTER TABLE agents ADD COLUMN reasoning_effort TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE agents DROP COLUMN reasoning_effort;
ALTER TABLE agents DROP COLUMN reasoning_effort_enabled;
