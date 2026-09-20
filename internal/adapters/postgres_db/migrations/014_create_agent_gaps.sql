-- +goose Up
ALTER TABLE agents ADD COLUMN gap_reporting_enabled BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE agent_gaps (
    id                   TEXT PRIMARY KEY,
    agent_id             TEXT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    category             TEXT NOT NULL CHECK(category IN ('knowledge', 'capability', 'improvement')),
    context              TEXT NOT NULL,
    details              TEXT NOT NULL,
    suggested_resolution TEXT,
    status               TEXT NOT NULL DEFAULT 'open' CHECK(status IN ('open', 'resolved', 'dismissed')),
    dismissal_category   TEXT CHECK(dismissal_category IN ('unrelated', 'insufficient_detail', 'duplicate', 'other')),
    dismissal_reason     TEXT,
    created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE agent_gap_references (
    id          TEXT PRIMARY KEY,
    gap_id      TEXT NOT NULL REFERENCES agent_gaps(id) ON DELETE CASCADE,
    session_id  TEXT REFERENCES playground_sessions(id) ON DELETE CASCADE,
    message_id  TEXT REFERENCES playground_messages(id) ON DELETE CASCADE,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_agent_gaps_agent ON agent_gaps(agent_id);
CREATE INDEX idx_agent_gap_references_gap ON agent_gap_references(gap_id);

-- +goose Down
DROP TABLE agent_gap_references;
DROP TABLE agent_gaps;
ALTER TABLE agents DROP COLUMN gap_reporting_enabled;
