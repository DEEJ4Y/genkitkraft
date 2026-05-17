-- +goose Up
ALTER TABLE agents ADD COLUMN max_tool_calls INTEGER NOT NULL DEFAULT 10;

-- +goose Down
ALTER TABLE agents DROP COLUMN max_tool_calls;
