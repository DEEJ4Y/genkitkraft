-- +goose Up
ALTER TABLE providers ADD COLUMN config TEXT;
UPDATE providers SET config = '{}' WHERE config IS NULL;
ALTER TABLE providers MODIFY config TEXT NOT NULL;
ALTER TABLE providers MODIFY api_key TEXT;

-- +goose Down
UPDATE providers SET api_key = '' WHERE api_key IS NULL;
ALTER TABLE providers MODIFY api_key TEXT NOT NULL;
ALTER TABLE providers DROP COLUMN config;
