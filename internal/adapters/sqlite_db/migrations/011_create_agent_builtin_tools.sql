-- +goose Up
CREATE TABLE agent_builtin_tools (
    agent_id       TEXT NOT NULL,
    builtin_tool_id TEXT NOT NULL,
    PRIMARY KEY (agent_id, builtin_tool_id),
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE agent_builtin_tools;
