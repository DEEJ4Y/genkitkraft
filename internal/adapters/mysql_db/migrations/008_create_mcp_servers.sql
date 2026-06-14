-- +goose Up
CREATE TABLE mcp_servers (
    id         VARCHAR(36) PRIMARY KEY,
    name       TEXT NOT NULL,
    transport  VARCHAR(20) NOT NULL DEFAULT 'sse',
    url        TEXT NOT NULL,
    headers    TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE mcp_servers;
