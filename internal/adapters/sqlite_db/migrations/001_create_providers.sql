-- +goose Up
CREATE TABLE providers (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    provider_type TEXT NOT NULL UNIQUE,
    api_key       TEXT,
    base_url      TEXT NOT NULL DEFAULT '',
    config        TEXT NOT NULL DEFAULT '{}',
    enabled       INTEGER NOT NULL DEFAULT 1,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE providers;
