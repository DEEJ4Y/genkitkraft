-- +goose Up
CREATE TABLE agents (
    id               VARCHAR(36) PRIMARY KEY,
    name             TEXT NOT NULL,
    provider_id      VARCHAR(36) NOT NULL,
    model_id         TEXT NOT NULL,
    system_prompt_id VARCHAR(36),
    temperature      DOUBLE NOT NULL DEFAULT 0.95,
    top_p            DOUBLE NOT NULL DEFAULT 0.95,
    top_k            INTEGER NOT NULL DEFAULT 40,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (provider_id) REFERENCES providers(id),
    FOREIGN KEY (system_prompt_id) REFERENCES prompts(id)
);

-- +goose Down
DROP TABLE agents;
