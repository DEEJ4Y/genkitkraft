---
sidebar_position: 3
---

# Managing Agents

Agents are the core concept in GenKitKraft — they combine a provider, model, and system prompt into a configured AI assistant that you can chat with or call via the API.

## Creating an Agent

1. Navigate to **Agents** in the sidebar.
2. Click **Create Agent**.
3. Fill in the form:
   - **Name** — A descriptive name for your agent.
   - **Provider** — Select a configured LLM provider.
   - **Model** — Choose a model from the selected provider. If the model doesn't exist in the dropdown, simply type the correct model name (e.g., `gpt-5-mini`) and it will be saved for future use.
   - **System Prompt** — Select a system prompt.
   - **Generation Parameters** (optional):
     - **Temperature** — Controls randomness (0 = deterministic, higher = more creative).
     - **Top P** — Nucleus sampling threshold.
     - **Top K** — Limits token selection to the top K candidates.
4. Click **Save**.

## Generation Parameters

| Parameter       | Range   | Description                                                                                                                                                |
| --------------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Temperature** | 0.0–2.0 | Controls randomness. Lower values produce more focused, deterministic output. Higher values increase creativity and variation. Default varies by provider. |
| **Top P**       | 0.0–1.0 | Controls diversity via nucleus sampling. The model considers only the smallest set of tokens whose cumulative probability exceeds this threshold.          |
| **Top K**       | Integer | Restricts the model's choices to the K most likely tokens at each step. Lower values make output more predictable.                                         |

## Configuring Tools

Each agent can be assigned specific tools — HTTP tools and MCP server tools — that it can use during conversations.

1. Open an agent for editing.
2. Click the **Tools** tab.
3. Add HTTP tools and MCP server tools as needed.
4. Click **Save Tools** to persist the configuration.

For detailed instructions on setting up tools and assigning them to agents, see the [Tools guide](./tools).

## Editing and Deleting Agents

- Click the **edit** icon on an agent card to modify its configuration.
- Click **delete** to remove an agent. This also removes any associated playground sessions.

## Using Agents

After creating an agent, you can:

- **Test it interactively** in the [Playground](./playground).
- **Deploy it** via the OpenAI-compatible API. Open the **Deploy** tab in the agent edit screen to find your agent ID, the full endpoint URL, and a ready-to-use curl command. See the [Deploy API documentation](../api/deploy) for the full reference.
