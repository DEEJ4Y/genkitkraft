-- +goose Up
ALTER TABLE agents ADD COLUMN temperature_enabled TINYINT(1) NOT NULL DEFAULT 0;
ALTER TABLE agents ADD COLUMN top_p_enabled TINYINT(1) NOT NULL DEFAULT 0;
ALTER TABLE agents ADD COLUMN top_k_enabled TINYINT(1) NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE agents DROP COLUMN temperature_enabled;
ALTER TABLE agents DROP COLUMN top_p_enabled;
ALTER TABLE agents DROP COLUMN top_k_enabled;
