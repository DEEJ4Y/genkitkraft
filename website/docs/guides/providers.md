---
sidebar_position: 1
---

# Configuring Providers

Providers are the LLM services that power your agents. GenKitKraft needs at least one configured provider before you can create agents. Each provider connects to a different AI service using your API key.

## Supported Providers

| Provider | Available Models |
|----------|-----------------|
| **Google AI** | gemini-3.1-pro-preview, gemini-3-flash-preview, gemini-3.1-flash-lite-preview, gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-lite |
| **Vertex AI** | Same models as Google AI (uses Google Cloud credentials instead of an API key) |
| **OpenAI** | gpt-5.4, gpt-5.4-mini, gpt-5.4-nano, gpt-5.3-codex, gpt-4o, gpt-4o-mini |
| **Anthropic** | claude-opus-4-6, claude-sonnet-4-6, claude-opus-4-5-20251202, claude-sonnet-4-5-20250929, claude-haiku-4-5-20251001 |
| **xAI** | grok-4.20-beta, grok-4-1-fast-reasoning, grok-4-1-fast-non-reasoning, grok-4, grok-code-fast-1 |
| **DeepSeek** | deepseek-chat, deepseek-reasoner |

## Adding a Provider

1. Navigate to **Settings** in the sidebar.
2. Click **Add Provider**.
3. Select the provider type from the dropdown.
4. Enter a display name and your API key.
5. Click **Save**.

:::note
API keys are encrypted at rest using the `ENCRYPTION_KEY` you configured during setup.
:::

## Testing a Provider

After adding a provider, click the **Test** button on the provider card. This verifies that the API key is valid and the provider is reachable.

## Editing and Deleting Providers

- Click the **edit** icon on a provider card to update its name or API key.
- Click the **delete** icon to remove a provider. Agents using this provider will no longer function until they are reconfigured with a different provider.
