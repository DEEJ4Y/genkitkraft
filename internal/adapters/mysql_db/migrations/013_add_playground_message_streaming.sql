-- +goose Up
ALTER TABLE playground_messages ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'complete' CHECK(status IN ('streaming', 'complete', 'error'));

CREATE TABLE playground_message_chunks (
    id          VARCHAR(36) PRIMARY KEY,
    message_id  VARCHAR(36) NOT NULL,
    seq         INTEGER NOT NULL,
    content     TEXT NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(message_id, seq),
    FOREIGN KEY (message_id) REFERENCES playground_messages(id) ON DELETE CASCADE
);

CREATE INDEX idx_playground_message_chunks_message ON playground_message_chunks(message_id);

-- +goose Down
DROP TABLE playground_message_chunks;
ALTER TABLE playground_messages DROP COLUMN status;
