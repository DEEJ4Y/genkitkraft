# Add fallback_session_id to playground MCP tools and update MCP docs

**Date:** 2026-10-09
**Branch:** fix/playground-fallback-session-id

## Summary

Some MCP clients drop the `session_id` argument. The server then rejects `playground_chat` and `playground_messages_list` with `missing properties: ["session_id"]`. Both tools now accept an optional `fallback_session_id`. The MCP docs now match the code.

## Changes

- `internal/handlers/mcp_handler/playground_tools.go`: Add `fallback_session_id` to `PlaygroundChatInput` and `ListPlaygroundMessagesInput`. Make `session_id` optional in the schema. Add `resolveSessionID`. It uses `fallback_session_id` first and then `session_id`. If neither is set, the tool returns an error before it saves a message. Update the tool descriptions.
- `internal/handlers/mcp_handler/playground_tools_test.go`: Add tests for `session_id` only, `fallback_session_id` only, both set, and neither set.
- `website/docs/guides/mcp-quickstart.md`: Add the gaps tools, `built_in_tools_list`, the `backup` and `restore` prompts, the chat overrides, and a troubleshooting entry for `session_id`.
- `website/docs/guides/mcp-agent-creation-guide.md`: Fix the chat field name (`content`, not `message`). Mark `title` as optional. Add the missing tools and the prompts to the copyable prompt.
- `website/docs/getting-started/mcp-quickstart.md`: Add gaps and built-in tools. Mention the `backup` and `restore` prompts.
- `website/docs/guides/gaps.md`: Add a link to the MCP gap tools.
- `internal/handlers/mcp_handler/prompts/create_agent.md`: Add the gaps tools, `built_in_tools_list`, and the `fallback_session_id` note.
- `.claude/skills/dev-flow/SKILL.md`: Add the rule "Keeping MCP Docs in Sync", a workflow step, and a checklist item. Add the missing files to the MCP file list.

## Notes

- If a client drops every argument with `session_id` in its name, `fallback_session_id` does not help. The live check with Claude desktop worked.
- `create_agent.md` is embedded in the binary. Rebuild the image to serve the new text.
- The machine used for this work has no C compiler. The SQLite tests in `internal/handlers/mcp_handler` did not run here. Checks on the live server passed.
