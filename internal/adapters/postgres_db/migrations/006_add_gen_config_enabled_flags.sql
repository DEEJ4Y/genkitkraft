-- +goose Up
ALTER TABLE agents ADD COLUMN temperature_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE agents ADD COLUMN top_p_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE agents ADD COLUMN top_k_enabled BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE agents DROP COLUMN temperature_enabled;
ALTER TABLE agents DROP COLUMN top_p_enabled;
ALTER TABLE agents DROP COLUMN top_k_enabled;
