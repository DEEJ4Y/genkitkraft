# Add reasoning effort option to agents

**Date:** 2026-10-11
**Branch:** feat/reasoning-effort

## Summary

An agent has a new option: reasoning effort. The option is off by default. When the option is off, requests do not change. When the option is on, the chat provider adds the setting that the provider uses for reasoning. The UI shows the values for each provider and accepts a typed value.

## Changes

- `spec/models/agent.tsp`: Add `reasoningEffortEnabled` and `reasoningEffort` to the agent response, create request, and update request.
- `spec/models/playground.tsp`: Add the same two fields as playground overrides.
- `spec/tsp-output/schema/openapi.yaml`, `internal/api/gen/types.gen.go`, `ui/lib/api/schema.d.ts`: Regenerate with `make generate`.
- `internal/domain/agent/entity.go`: Add the two fields to `Agent`. Add `MaxReasoningEffortLength`.
- `internal/ports/chat_provider/port.go`: Add the two fields to `ChatRequest`.
- `internal/adapters/sqlite_db/migrations/015_add_reasoning_effort.sql`, `internal/adapters/mysql_db/migrations/015_add_reasoning_effort.sql`, `internal/adapters/postgres_db/migrations/015_add_reasoning_effort.sql`: Add the `reasoning_effort_enabled` and `reasoning_effort` columns.
- `internal/adapters/sqlite_agent/repository.go`, `internal/adapters/mysql_agent/repository.go`, `internal/adapters/postgres_agent/repository.go`: Read and write the new columns.
- `internal/app/commands/reasoning_effort.go`: New helper that trims and checks the value.
- `internal/app/commands/create_agent.go`, `internal/app/commands/update_agent.go`: Accept and store the new fields.
- `internal/app/queries/resolve_playground_config.go`: Resolve the new fields and the playground overrides into `ChatRequest`.
- `internal/adapters/genkit_chat_provider/provider.go`: `buildConfig` returns an error. Each config builder sets the provider reasoning setting when the option is on. The Google and Ollama builders now return an untyped `nil` when no setting exists. The Anthropic builder sets `MaxTokens` to the budget plus 4096 and does not send temperature, top P, or top K while thinking is on.
- `internal/adapters/genkit_chat_provider/reasoning_effort.go`: New helpers that map the value to the Google, Anthropic, and Ollama settings.
- `internal/handlers/http_handler/type_conversion.go`, `internal/handlers/http_handler/handler.go`: Map the new fields.
- `internal/handlers/mcp_handler/agent_tools.go`, `internal/handlers/mcp_handler/playground_tools.go`: Add `reasoning_effort_enabled` and `reasoning_effort` to the tool inputs and outputs.
- `internal/handlers/mcp_handler/prompts/create_agent.md`, `backup.md`, `restore.md`: Describe the option.
- `ui/lib/reasoning-options.ts`: New list of values for each provider type, and a check for the value.
- `ui/components/ReasoningEffortField.tsx`: New toggle and dropdown that accepts a typed value.
- `ui/components/AgentForm.tsx`, `ui/components/AgentCard.tsx`: Show and save the option.
- `ui/components/playground/PlaygroundConfigBar.tsx`, `ui/components/playground/AgentPlayground.tsx`, `ui/hooks/usePlaygroundChat.ts`: Add the option as a playground override and to Save to Agent.
- `website/docs/guides/agents.md`, `website/docs/guides/playground.md`, `website/docs/guides/mcp-quickstart.md`, `website/docs/guides/mcp-agent-creation-guide.md`: Describe the option.
- `internal/adapters/genkit_chat_provider/config_test.go`, `internal/app/commands/agent_reasoning_effort_test.go`, `internal/app/queries/resolve_playground_config_test.go`: Add tests.

## Notes

- Azure AI Foundry and Bedrock do not support the option. Their plugins ignore a typed config. The UI hides the toggle for them.
- No C compiler is available on the local machine, so `CGO_ENABLED=0` there. The integration tests ran in a `golang:1.25` Docker container with `go test -tags integration ./...`.
- All packages passed. Four MySQL tests timed out at first because many MySQL containers started in parallel. They passed on a rerun with `-p 1`.
- The existing MySQL and PostgreSQL agent repository tests cover the new columns in the create and update round trips.
- `website/docs/api/endpoints.md` was not changed. Its request examples use an old shape (`generationConfig`).
- Anthropic thinking needs a budget of at least 1024 tokens. The model must support thinking, or the provider returns an error.
