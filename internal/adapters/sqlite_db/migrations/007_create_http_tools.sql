-- +goose Up
CREATE TABLE http_tools (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    method        TEXT NOT NULL DEFAULT 'GET',
    url           TEXT NOT NULL,
    headers       TEXT NOT NULL DEFAULT '[]',
    body_template TEXT NOT NULL DEFAULT '',
    input_schema  TEXT NOT NULL DEFAULT '{}',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE http_tools;
