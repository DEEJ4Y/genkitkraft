-- +goose Up
ALTER TABLE agents ADD COLUMN gap_reporting_enabled TINYINT(1) NOT NULL DEFAULT 0;

CREATE TABLE agent_gaps (
    id                   VARCHAR(36) PRIMARY KEY,
    agent_id             VARCHAR(36) NOT NULL,
    category             VARCHAR(20) NOT NULL CHECK(category IN ('knowledge', 'capability', 'improvement')),
    context              TEXT NOT NULL,
    details              TEXT NOT NULL,
    suggested_resolution TEXT,
    status               VARCHAR(20) NOT NULL DEFAULT 'open' CHECK(status IN ('open', 'resolved', 'dismissed')),
    dismissal_category   VARCHAR(30) CHECK(dismissal_category IN ('unrelated', 'insufficient_detail', 'duplicate', 'other')),
    dismissal_reason     TEXT,
    created_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

CREATE TABLE agent_gap_references (
    id          VARCHAR(36) PRIMARY KEY,
    gap_id      VARCHAR(36) NOT NULL,
    session_id  VARCHAR(36),
    message_id  VARCHAR(36),
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (gap_id) REFERENCES agent_gaps(id) ON DELETE CASCADE,
    FOREIGN KEY (session_id) REFERENCES playground_sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (message_id) REFERENCES playground_messages(id) ON DELETE CASCADE
);

CREATE INDEX idx_agent_gaps_agent ON agent_gaps(agent_id);
CREATE INDEX idx_agent_gap_references_gap ON agent_gap_references(gap_id);

-- +goose Down
DROP TABLE agent_gap_references;
DROP TABLE agent_gaps;
ALTER TABLE agents DROP COLUMN gap_reporting_enabled;
