namespace Api.Models;

/** Configuration of an MCP server's tool selection for an agent. */
model AgentMcpServerToolConfig {
  /** ID of the MCP server. */
  mcpServerId: string;

  /** Whether all tools from this server are selected. When true, toolNames is ignored. */
  selectAll: boolean;

  /** Specific tool names selected from this server (used when selectAll is false). */
  toolNames: string[];
}

/** Response containing an agent's tool configuration. */
model AgentToolConfigResponse {
  /** IDs of HTTP tools assigned to this agent. */
  httpToolIds: string[];

  /** MCP server tool configurations for this agent. */
  mcpServers: AgentMcpServerToolConfig[];

  /** IDs of built-in tools enabled for this agent. */
  builtInToolIds: string[];
}

/** Request to update an agent's tool configuration. */
model UpdateAgentToolConfigRequest {
  /** IDs of HTTP tools to assign. */
  httpToolIds: string[];

  /** MCP server tool configurations to assign. */
  mcpServers: AgentMcpServerToolConfig[];

  /** IDs of built-in tools to enable. */
  builtInToolIds: string[];
}

/** MCP server tool override for playground chat. */
model PlaygroundMcpServerToolOverride {
  /** ID of the MCP server. */
  mcpServerId: string;

  /** Whether all tools from this server are selected. */
  selectAll: boolean;

  /** Specific tool names (used when selectAll is false). */
  toolNames: string[];
}
