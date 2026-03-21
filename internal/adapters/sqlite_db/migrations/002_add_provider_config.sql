-- +goose Up
-- Add config column and make api_key nullable.
-- SQLite requires table recreation to change column constraints.

CREATE TABLE providers_new (
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

INSERT INTO providers_new (id, name, provider_type, api_key, base_url, enabled, created_at, updated_at)
SELECT id, name, provider_type, api_key, base_url, enabled, created_at, updated_at
FROM providers;

DROP TABLE providers;

ALTER TABLE providers_new RENAME TO providers;

-- +goose Down
CREATE TABLE providers_old (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    provider_type TEXT NOT NULL UNIQUE,
    api_key       TEXT NOT NULL,
    base_url      TEXT NOT NULL DEFAULT '',
    enabled       INTEGER NOT NULL DEFAULT 1,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO providers_old (id, name, provider_type, api_key, base_url, enabled, created_at, updated_at)
SELECT id, name, provider_type, COALESCE(api_key, ''), base_url, enabled, created_at, updated_at
FROM providers;

DROP TABLE providers;

ALTER TABLE providers_old RENAME TO providers;
