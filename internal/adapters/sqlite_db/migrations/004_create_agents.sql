-- +goose Up
CREATE TABLE agents (
    id               TEXT PRIMARY KEY,
    name             TEXT NOT NULL,
    provider_id      TEXT NOT NULL REFERENCES providers(id),
    model_id         TEXT NOT NULL,
    system_prompt_id TEXT REFERENCES prompts(id),
    temperature      REAL NOT NULL DEFAULT 0.95,
    top_p            REAL NOT NULL DEFAULT 0.95,
    top_k            INTEGER NOT NULL DEFAULT 40,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE agents;
