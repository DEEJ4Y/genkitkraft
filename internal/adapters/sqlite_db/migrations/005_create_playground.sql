-- +goose Up
CREATE TABLE playground_sessions (
    id          TEXT PRIMARY KEY,
    agent_id    TEXT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    title       TEXT NOT NULL DEFAULT 'New Session',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE playground_messages (
    id          TEXT PRIMARY KEY,
    session_id  TEXT NOT NULL REFERENCES playground_sessions(id) ON DELETE CASCADE,
    role        TEXT NOT NULL CHECK(role IN ('user', 'assistant')),
    content     TEXT NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_playground_sessions_agent ON playground_sessions(agent_id);
CREATE INDEX idx_playground_messages_session ON playground_messages(session_id);

-- +goose Down
DROP TABLE playground_messages;
DROP TABLE playground_sessions;
