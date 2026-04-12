---
sidebar_position: 2
---

# First Steps

After installing GenKitKraft, follow these steps to configure your first agent.

## 1. Open the UI

Navigate to [http://localhost:8080](http://localhost:8080) in your browser. You'll see the GenKitKraft dashboard.

## 2. Configure a Provider

Go to **Settings** and add at least one LLM provider (e.g. OpenAI, Anthropic, or Google). You'll need an API key from your chosen provider. See the [Providers guide](../guides/providers) for details.

## 3. Create a System Prompt

Go to **Prompts** and create a system prompt that defines your agent's behavior and personality. See the [Prompts guide](../guides/prompts) for more information.

## 4. Create an Agent

Go to **Agents** and create a new agent by selecting a provider, model, and system prompt. This ties everything together into a usable agent. See the [Agents guide](../guides/agents) for details.

## 5. Test in the Playground

Click on your agent to open the built-in playground. Use it to chat with your agent and verify it behaves as expected. See the [Playground guide](../guides/playground) for more.

## 6. Optional: Enable Authentication

If you're exposing GenKitKraft to a network, set the `AUTH_CREDENTIALS` environment variable to secure your instance with login credentials. See [Authentication](../configuration/authentication) for setup instructions.
