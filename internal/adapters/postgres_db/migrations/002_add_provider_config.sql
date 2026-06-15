-- +goose Up
ALTER TABLE providers ADD COLUMN config TEXT NOT NULL DEFAULT '{}';
ALTER TABLE providers ALTER COLUMN api_key DROP NOT NULL;

-- +goose Down
UPDATE providers SET api_key = '' WHERE api_key IS NULL;
ALTER TABLE providers ALTER COLUMN api_key SET NOT NULL;
ALTER TABLE providers DROP COLUMN config;
