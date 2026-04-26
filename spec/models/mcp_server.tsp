namespace Api.Models;

/** Transport type for an MCP server connection. */
enum McpTransport {
  SSE: "sse",
  StreamableHttp: "streamableHttp",
}

/** A key-value pair representing an HTTP header for MCP server connections. */
model McpServerHeader {
  /** Header name (e.g. Authorization). */
  name: string;

  /** Header value. Supports Go template expressions (e.g. Bearer {{.token}}). */
  value: string;
}

/** A remote MCP server configuration. */
model McpServerResponse {
  /** Unique MCP server ID. */
  id: string;

  /** Display name for this MCP server. */
  name: string;

  /** Transport type used to connect to the MCP server. */
  transport: McpTransport;

  /** URL of the MCP server endpoint. */
  url: string;

  /** HTTP headers to include when connecting to the MCP server. */
  headers: McpServerHeader[];

  /** When this MCP server was created. */
  createdAt: utcDateTime;

  /** When this MCP server was last updated. */
  updatedAt: utcDateTime;
}

/** Request to create a new MCP server configuration. */
model CreateMcpServerRequest {
  /** Display name for this MCP server. */
  name: string;

  /** Transport type used to connect to the MCP server. */
  transport: McpTransport;

  /** URL of the MCP server endpoint. */
  url: string;

  /** HTTP headers to include when connecting. */
  headers?: McpServerHeader[];
}

/** Request to update an existing MCP server configuration. */
model UpdateMcpServerRequest {
  /** Updated display name. */
  name?: string;

  /** Updated transport type. */
  transport?: McpTransport;

  /** Updated URL. */
  url?: string;

  /** Updated HTTP headers. */
  headers?: McpServerHeader[];
}

/** Paginated list of MCP servers. */
model McpServerListResponse {
  /** Array of MCP servers. */
  mcpServers: McpServerResponse[];

  /** Total number of MCP servers. */
  total: int32;

  /** Number of MCP servers per page. */
  limit: int32;

  /** Number of MCP servers skipped. */
  offset: int32;
}

/** A tool available from an MCP server. */
model McpServerToolResponse {
  /** Tool name as reported by the MCP server. */
  name: string;

  /** Tool description as reported by the MCP server. */
  description: string;
}

/** List of tools available from an MCP server. */
model McpServerToolListResponse {
  /** Array of available tools. */
  tools: McpServerToolResponse[];
}
