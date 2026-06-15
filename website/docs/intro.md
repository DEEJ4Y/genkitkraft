---
sidebar_position: 1
slug: /
---

# Introduction

GenKitKraft is a self-hostable platform for configuring and running LLM agents, built on [Google Genkit](https://github.com/firebase/genkit) (Go SDK). It ships as a single binary with an embedded frontend and SQLite storage — no external dependencies required.

## Features

- **Agent Builder UI** — Create agents with custom system instructions, model selection, temperature, and tool call limits
- **Built-in Tools** — `web_fetch` fetches any URL and converts it to Markdown; no external setup required
- **MCP Server** — Manage everything via Model Context Protocol from Claude Desktop, Cursor, or any MCP client
- **MCP Tool Support** — Connect MCP tool servers to your agents for external integrations
- **Multi-Provider LLM Access** — Use OpenAI, Anthropic, Google, and more
- **OpenAI-Compatible API** — Expose configured agents via a standard API
- **Flexible Storage** — SQLite by default (zero external dependencies); PostgreSQL, MySQL, and MariaDB supported for multi-instance deployments

## Getting Started

There are two ways to get started with GenKitKraft:

### Via the UI

1. [Install GenKitKraft](/docs/getting-started/installation)
2. Follow the [First Steps](/docs/getting-started/first-steps) guide to configure your first agent through the web interface

### Via MCP (Claude Desktop, Cursor, etc.)

1. [Install GenKitKraft](/docs/getting-started/installation)
2. Follow the [MCP Quickstart](/docs/getting-started/mcp-quickstart) to connect your MCP client and create agents conversationally

## Configuration

GenKitKraft is configured entirely through environment variables. For configuration details, see [Environment Variables](/docs/configuration/environment-variables).

## Using the UI

To learn about the UI, start with [Configuring Providers](/docs/guides/providers). Once you have agents set up, see the [Tools guide](/docs/guides/tools) to connect HTTP tools and MCP servers.

## Using the MCP Server

GenKitKraft exposes all management APIs as MCP tools. Connect from any MCP-compatible client to create agents, configure providers, assign tools, and chat — all conversationally. See the [MCP Quickstart](/docs/getting-started/mcp-quickstart) to get connected, or the [full MCP reference](/docs/guides/mcp-quickstart) for all available tools.
