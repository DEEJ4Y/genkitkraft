-- +goose Up
CREATE TABLE http_tools (
    id            VARCHAR(36) PRIMARY KEY,
    name          TEXT NOT NULL,
    description   TEXT NOT NULL,
    method        VARCHAR(10) NOT NULL DEFAULT 'GET',
    url           TEXT NOT NULL,
    headers       TEXT NOT NULL,
    body_template TEXT NOT NULL,
    input_schema  TEXT NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE http_tools;
