-- +goose Up
CREATE TABLE agent_builtin_tools (
    agent_id        VARCHAR(36) NOT NULL,
    builtin_tool_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (agent_id, builtin_tool_id),
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE agent_builtin_tools;
