---
sidebar_position: 4
---

# Using the Playground

The playground is a built-in chat interface for testing your agents interactively. It supports real-time streaming responses and lets you experiment with different configurations without modifying the agent itself.

## Starting a Chat

1. Navigate to **Agents** and click on an agent.
2. The playground opens with the agent's configuration loaded.
3. Type a message and press **Enter** or click **Send**.
4. Responses stream in real-time via Server-Sent Events (SSE).

## Managing Sessions

- **Create a new session** — Click the **New Session** button in the sidebar to start a fresh conversation.
- **Switch between sessions** — Click on a session in the sidebar to load its history.
- **Delete a session** — Click the delete icon on a session to remove it and its messages.

Sessions are preserved across page reloads, so you can pick up where you left off.

## Configuration Overrides

In the playground config bar, you can temporarily override:

- **Provider** — Test with a different LLM provider.
- **Model** — Try a different model. If the model doesn't exist in the dropdown, simply type the correct model name (e.g., `gpt-5-mini`) and it will be saved for future use.
- **Temperature, Top P, Top K** — Experiment with generation parameters.

These overrides apply only to the current playground session and don't modify the agent's saved configuration.

## Saving a Configuration

If you find a combination of overrides that works well, click the **Save Config** button to save it as a new agent. This lets you quickly iterate on configurations and promote the best ones to reusable agents.
