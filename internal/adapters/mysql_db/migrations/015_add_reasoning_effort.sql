-- +goose Up
ALTER TABLE agents ADD COLUMN reasoning_effort_enabled TINYINT(1) NOT NULL DEFAULT 0;
ALTER TABLE agents ADD COLUMN reasoning_effort VARCHAR(64) NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE agents DROP COLUMN reasoning_effort;
ALTER TABLE agents DROP COLUMN reasoning_effort_enabled;
