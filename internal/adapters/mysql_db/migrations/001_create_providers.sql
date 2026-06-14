-- +goose Up
CREATE TABLE providers (
    id            VARCHAR(36) PRIMARY KEY,
    name          TEXT NOT NULL,
    provider_type VARCHAR(255) NOT NULL UNIQUE,
    api_key       TEXT NOT NULL,
    base_url      VARCHAR(2048) NOT NULL DEFAULT '',
    enabled       TINYINT(1) NOT NULL DEFAULT 1,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE providers;
