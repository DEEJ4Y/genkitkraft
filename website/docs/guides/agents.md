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
     - **Reasoning Effort** — Off by default. Turn it on to set how much the model reasons (thinking). See [Reasoning Effort](#reasoning-effort).
     - **Max Tool Calls** — Maximum number of tool call iterations per request (default 10).
4. Click **Save**.

## Generation Parameters

| Parameter          | Range   | Description                                                                                                                                                |
| ------------------ | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Temperature**    | 0.0–2.0 | Controls randomness. Lower values produce more focused, deterministic output. Higher values increase creativity and variation. Default varies by provider. |
| **Top P**          | 0.0–1.0 | Controls diversity via nucleus sampling. The model considers only the smallest set of tokens whose cumulative probability exceeds this threshold.          |
| **Top K**          | Integer | Restricts the model's choices to the K most likely tokens at each step. Lower values make output more predictable.                                         |
| **Max Tool Calls** | Integer | Limits how many tool call iterations the agent can perform in a single request. Prevents runaway loops when tools trigger additional tool calls. Default is 10. |

## Reasoning Effort

Some models can reason (think) before they answer. The **Reasoning Effort** option sets how much. It is off by default. When it is off, GenKitKraft sends the same request as before. When it is on, GenKitKraft adds the setting that the provider uses for reasoning.

1. Turn on the **Reasoning Effort** toggle.
2. Select a value from the list, or type your own value. The list shows the values for the provider of the agent. A typed value is sent to the provider as it is, so you can use new values when a provider adds them.

| Provider | Setting that is sent | Values in the list |
| --- | --- | --- |
| Google AI, Vertex AI | Thinking level, or thinking budget when the value is a number | `minimal`, `low`, `medium`, `high`, `-1` (dynamic) |
| Anthropic | Thinking budget in tokens | `1024`, `2048`, `4096`, `8192`, `16384` |
| Ollama | `think` | `true`, `false`, `low`, `medium`, `high` |
| OpenAI, Azure OpenAI, xAI, DeepSeek, OpenAI-compatible | `reasoning_effort` | `minimal`, `low`, `medium`, `high` |
| Azure AI Foundry, Bedrock | Not supported | — |

:::note
- The model must support reasoning. If it does not, the provider returns an error.
- For Anthropic, the value must be a whole number of at least 1024. While reasoning effort is on, GenKitKraft does not send Temperature, Top P, or Top K, because Anthropic does not accept them with thinking.
- For Gemini 3 models, use a level. For Gemini 2.5 models, use a token budget.
:::

You can also change reasoning effort for one playground session. See the [Playground guide](./playground).

## Configuring Tools

Each agent can be assigned specific tools — HTTP tools and MCP server tools — that it can use during conversations.

1. Open an agent for editing.
2. Click the **Tools** tab.
3. Add HTTP tools and MCP server tools as needed.
4. Click **Save Tools** to persist the configuration.

For detailed instructions on setting up tools and assigning them to agents, see the [Tools guide](./tools).

## Gap Reporting

Each agent can also self-report gaps it notices during a conversation — questions it couldn't answer, actions it couldn't perform, or ideas to automate more of the flow. Enable it via the **Enable Gap Reporting** toggle on the Configuration tab, then review reported gaps on the agent's **Gaps** tab. See the [Gaps guide](./gaps) for details.

## Editing and Deleting Agents

- Click the **edit** icon on an agent card to modify its configuration.
- Click **delete** to remove an agent. This also removes any associated playground sessions.

## Using Agents

After creating an agent, you can:

- **Test it interactively** in the [Playground](./playground).
- **Deploy it** via the OpenAI-compatible API. Open the **Deploy** tab in the agent edit screen to find your agent ID, the full endpoint URL, and a ready-to-use curl command. See the [Deploy API documentation](../api/deploy) for the full reference.
- **Embed it in your website** with the chat widget. Open the **Widget** tab to design it, then follow the [chat widget guide](./chat-widget).
