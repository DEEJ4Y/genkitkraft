---
sidebar_position: 5
---

# Tools

Tools give your agents the ability to take actions beyond generating text — call APIs, query databases, interact with external services, and more. GenKitKraft supports three types of tools: **Built-in Tools**, **HTTP Tools**, and **MCP Servers**.

## Built-in Tools

Built-in tools are provided by GenKitKraft itself — no external configuration required. They run inside the server and are always available to enable on any agent.

### Web Fetch

The **Web Fetch** tool lets your agent fetch content from any URL and receive it as Markdown. This is useful for browsing documentation, referencing web pages, or pulling in context from external sources during a conversation.

**Capabilities:**
- Fetches static HTML content from any URL.
- Converts the HTML to Markdown for easy consumption by the LLM.

**Limitations:**
- Only supports static HTML content. It does not execute JavaScript or render dynamic content (no headless browser).
- Pages that require client-side rendering will return incomplete or empty results.

### Enabling Built-in Tools

1. Open an agent for editing.
2. Click the **Tools** tab.
3. In the **Built-in Tools** section, check the tools you want to enable.
4. Click **Save Tools**.

You can also enable built-in tools via the API by including `builtInToolIds` in the agent tool configuration request (e.g., `["web_fetch"]`).

## HTTP Tools

HTTP tools are simple, single-endpoint tools. Each tool makes one HTTP request and returns the response to the agent.

### Creating an HTTP Tool

1. Navigate to **Tools** in the sidebar.
2. On the **HTTP Tools** tab, click **Create HTTP Tool**.
3. Fill in the form:
   - **Name** — A descriptive name (e.g., "Get Weather").
   - **Description** — Explain what the tool does. The LLM reads this to decide when to use it.
   - **Method** — HTTP method (GET, POST, PUT, DELETE, PATCH).
   - **URL** — The full endpoint URL.
   - **Headers** (optional) — Add request headers as key-value pairs (e.g., `Authorization: Bearer ...`).
   - **Body** (optional) — Request body for POST/PUT/PATCH methods.
   - **Input Schema** (optional) — A JSON Schema describing the parameters the LLM should provide. When specified, the LLM will fill in parameters according to this schema and they will be included in the request.
4. Click **Save**.

### Editing and Deleting

- Click the **edit** icon on a tool card to modify it.
- Click **delete** to remove a tool. Any agents using this tool will lose access to it.

:::tip Docker networking
If your HTTP tool targets a service running on your host machine (e.g., `localhost:9090`), and GenKitKraft runs inside Docker, use `host.docker.internal` instead of `localhost` or `127.0.0.1`. Docker containers can't reach the host's localhost directly.
:::

## MCP Servers

[Model Context Protocol (MCP)](https://modelcontextprotocol.io/) servers expose collections of tools over a standardized protocol. A single MCP server can provide many tools. GenKitKraft connects to remote MCP servers and discovers their tools automatically.

### Transport Modes

MCP servers communicate over one of two transports:

| Transport | Description |
|-----------|-------------|
| **SSE** | Server-Sent Events — the original MCP transport. Uses a long-lived HTTP connection for bidirectional messaging. |
| **Streamable HTTP** | The newer transport (March 2025 spec). Uses standard HTTP POST requests. This is the recommended transport for new servers. |

GenKitKraft supports both transports and includes automatic fallback — if the configured transport fails, it tries the other one.

### Adding an MCP Server

1. Navigate to **Tools** in the sidebar.
2. Switch to the **MCP Servers** tab.
3. Click **Add MCP Server**.
4. Fill in the form:
   - **Name** — A descriptive name (e.g., "Todo Server", "GitHub Tools").
   - **Transport** — Select SSE or Streamable HTTP.
   - **URL** — The MCP server endpoint URL.
   - **Headers** (optional) — Add custom headers as key-value pairs (e.g., for authentication).
5. Click **Save**.

### Discovering Tools

After adding an MCP server, you can view the tools it provides:

- Click **Discover Tools** on the MCP server card.
- GenKitKraft connects to the server and lists all available tools with their names, descriptions, and input schemas.
- This is the same list you'll see when configuring tools for an agent.

:::tip Docker networking
The same Docker networking rules apply to MCP servers. Use `host.docker.internal` for services running on the host machine.
:::

## Agent Tool Configuration

Once you've configured built-in tools, HTTP tools, and MCP servers, you can assign specific tools to each agent.

### Tools Tab

1. Open an agent for editing.
2. Click the **Tools** tab.
3. You'll see three sections:

#### Enabling Built-in Tools

Check the built-in tools you want the agent to use (e.g., **Web Fetch**). These are always available and require no external setup.

#### Adding HTTP Tools

1. Click **Add HTTP Tool**.
2. A modal shows all available HTTP tools.
3. Select one or more tools to add.
4. Click **Add** — the selected tools appear in the HTTP Tools section.

#### Adding MCP Server Tools

1. Click **Add MCP Server**.
2. A modal shows all configured MCP servers.
3. Select a server — GenKitKraft connects to it and discovers its tools.
4. Each MCP server gets its own section listing all tools from that server.
5. Use checkboxes to select specific tools, or use **Select All** to include every tool.
6. You can add multiple MCP servers, each in its own section.

#### Saving

Click **Save Tools** to persist the tool configuration. The agent will have access to only the tools you've selected when used in the playground or via the API.

#### Removing

- Remove individual HTTP tools or entire MCP server sections using the delete icon.
- Uncheck individual MCP tools to exclude them while keeping the server connected.

## Playground Tool Overrides

The playground lets you temporarily override an agent's tool configuration without modifying the saved agent.

### Using Tool Overrides

1. Open an agent in the playground.
2. Click the **Tools** icon in the config bar (next to the configuration button).
3. A panel opens showing:
   - **Built-in Tools** — Available built-in tools with checkboxes. Tools saved in the agent's config are pre-selected.
   - **HTTP Tools** — All available HTTP tools with checkboxes. Tools saved in the agent's config are pre-selected.
   - **MCP Servers** — All configured MCP servers. Select a server to see its tools. Tools saved in the agent's config are pre-selected.
4. Check or uncheck tools to customize what the agent can use for this session.
5. The config bar shows a modified indicator when tool overrides differ from the saved config.

### Saving Overrides

Tool overrides are temporary by default. To persist them:

1. Click **Save Config** in the playground.
2. The save modal includes both configuration overrides (provider, model, parameters) and tool overrides.
3. Save to update the agent's configuration.

:::note
Tool overrides in the playground do not automatically save. This lets you experiment freely without accidentally modifying the agent's production configuration.
:::
