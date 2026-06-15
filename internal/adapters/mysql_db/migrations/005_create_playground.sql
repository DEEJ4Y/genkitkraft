-- +goose Up
CREATE TABLE playground_sessions (
    id          VARCHAR(36) PRIMARY KEY,
    agent_id    VARCHAR(36) NOT NULL,
    title       VARCHAR(255) NOT NULL DEFAULT 'New Session',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

CREATE TABLE playground_messages (
    id          VARCHAR(36) PRIMARY KEY,
    session_id  VARCHAR(36) NOT NULL,
    role        VARCHAR(20) NOT NULL CHECK(role IN ('user', 'assistant')),
    content     TEXT NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES playground_sessions(id) ON DELETE CASCADE
);

CREATE INDEX idx_playground_sessions_agent ON playground_sessions(agent_id);
CREATE INDEX idx_playground_messages_session ON playground_messages(session_id);

-- +goose Down
DROP TABLE playground_messages;
DROP TABLE playground_sessions;
