---
sidebar_position: 1
slug: /
---

# Introduction

GenKitKraft is a self-hostable platform for configuring and running LLM agents, built on [Google Genkit](https://github.com/firebase/genkit) (Go SDK). It ships as a single binary with an embedded frontend and SQLite storage — no external dependencies required.

## Features

- **Agent Builder UI** — Create agents with custom system instructions
- **MCP Tool Support** — Connect Model Context Protocol tool servers to your agents
- **Multi-Provider LLM Access** — Use OpenAI, Anthropic, Google, and more
- **OpenAI-Compatible API** — Expose configured agents via a standard API
- **Single Binary Deployment** — Server and UI ship as one binary

## Getting Started

For installation instructions, see [Installation](/docs/getting-started/installation).

Once running, follow the [First Steps](/docs/getting-started/first-steps) guide to configure your first agent.

## Configuration

GenKitKraft is configured entirely through environment variables. For configuration details, see [Environment Variables](/docs/configuration/environment-variables).

## Using the UI

To learn about the UI, start with [Configuring Providers](/docs/guides/providers).
