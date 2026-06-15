-- +goose Up
CREATE TABLE agent_http_tools (
    agent_id     TEXT NOT NULL,
    http_tool_id TEXT NOT NULL,
    PRIMARY KEY (agent_id, http_tool_id),
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE,
    FOREIGN KEY (http_tool_id) REFERENCES http_tools(id) ON DELETE CASCADE
);

CREATE TABLE agent_mcp_servers (
    agent_id      TEXT NOT NULL,
    mcp_server_id TEXT NOT NULL,
    select_all    BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (agent_id, mcp_server_id),
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE,
    FOREIGN KEY (mcp_server_id) REFERENCES mcp_servers(id) ON DELETE CASCADE
);

CREATE TABLE agent_mcp_server_tools (
    agent_id      TEXT NOT NULL,
    mcp_server_id TEXT NOT NULL,
    tool_name     TEXT NOT NULL,
    PRIMARY KEY (agent_id, mcp_server_id, tool_name),
    FOREIGN KEY (agent_id, mcp_server_id) REFERENCES agent_mcp_servers(agent_id, mcp_server_id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE agent_mcp_server_tools;
DROP TABLE agent_mcp_servers;
DROP TABLE agent_http_tools;
