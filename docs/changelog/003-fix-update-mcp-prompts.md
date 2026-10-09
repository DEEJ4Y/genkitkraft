# Update the MCP prompts for the latest capabilities

**Date:** 2026-10-09
**Branch:** fix/56-update-mcp-prompts
**Issue:** #56

## Summary

The `create-agent`, `backup`, and `restore` MCP prompts did not describe new capabilities. A restore also removed the built-in tools of an agent, because `agent_tools_update` replaces the whole configuration. The prompts now match the code. The `dev-flow` skill now has a rule to update the prompts when a capability changes.

## Changes

- `internal/handlers/mcp_handler/prompts/create_agent.md`: Add `max_tool_calls` and `gap_reporting_enabled` to the agent fields. Add a Gaps section (categories, status, `gaps_*` tools). Add built-in tools (`web_fetch`) as the third tool type. Add a warning that `agent_tools_update` replaces the whole configuration. Add the stream reconnect and cancel routes for the Deploy API and the Playground. Add the `playground_chat` overrides and a Chat Widget section. Fix the transport value to `streamableHttp`. Change the example workflow and the steps to create an agent.
- `internal/handlers/mcp_handler/prompts/backup.md`: Add `max_tool_calls`, `gap_reporting_enabled`, and built-in tools to the output format and the steps. Add `built_in_tools_list` to the tool table. State that conversation histories and gaps are not backed up.
- `internal/handlers/mcp_handler/prompts/restore.md`: Pass `max_tool_calls` and `gap_reporting_enabled` to `agents_create`. Pass `built_in_tool_ids` to `agent_tools_update`. Add the `*_update` tools to the tool table. Add warnings to the report. State that conversation histories and gaps are not restored.
- `.claude/skills/dev-flow/SKILL.md`: Add the section "Keeping MCP Prompts in Sync" (change-to-prompt table and the backup and restore rules). Add a workflow step and two checklist items (4b and 20a).

## Notes

- Backup and restore cover configuration only. Conversation histories and reported gaps are out of scope.
- The prompts are embedded in the binary. Rebuild the image to serve the new text.
- Go build and vet passed. Four SQLite tests in `internal/handlers/mcp_handler` fail on this machine because it has no C compiler (`CGO_ENABLED=0`). The changes do not touch Go code.
- No live backup and restore dry run was done in this change.
