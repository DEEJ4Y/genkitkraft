-- +goose Up
ALTER TABLE playground_messages ADD COLUMN status TEXT NOT NULL DEFAULT 'complete' CHECK(status IN ('streaming', 'complete', 'error'));

CREATE TABLE playground_message_chunks (
    id          TEXT PRIMARY KEY,
    message_id  TEXT NOT NULL REFERENCES playground_messages(id) ON DELETE CASCADE,
    seq         INTEGER NOT NULL,
    content     TEXT NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(message_id, seq)
);

CREATE INDEX idx_playground_message_chunks_message ON playground_message_chunks(message_id);

-- +goose Down
DROP TABLE playground_message_chunks;
ALTER TABLE playground_messages DROP COLUMN status;
