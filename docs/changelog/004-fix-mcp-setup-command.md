# Fix the Claude Desktop MCP setup command in the docs

**Date:** 2026-10-09
**Branch:** fix/mcp-setup-command
**Issue:** none

## Summary

The home page and the docs showed a Claude Desktop config with a bare `"url"` entry. Claude Desktop does not accept this entry, so the setup did not work. The README already used the correct form. The home page and the docs now use the same form as the README: `npx` with `mcp-remote`.

## Changes

- `website/src/pages/index.tsx`: Change `mcpConfigSnippet` to `command: npx` and `args: ["mcp-remote", ...]`. Change the label to "Claude Desktop config", because Cursor uses a different config.
- `website/docs/getting-started/mcp-quickstart.md`: Change the Claude Desktop config. Add a note that Node.js is necessary. Show the Basic Auth header for Claude Desktop (`--header` and `env`) and for URL clients (Cursor).
- `website/docs/guides/mcp-quickstart.md`: Change the Claude Desktop config and its auth example in the same way. Add the `headers` example to the Cursor section.
- `internal/handlers/mcp_handler/prompts/create_agent.md`: Show the Claude Desktop auth config with `mcp-remote`. Keep the `url` and `headers` form for Cursor and other URL clients.

## Notes

- The Cursor sections are unchanged. Cursor supports a URL.
- The `mcp-remote` header syntax comes from the `mcp-remote` README. It was not tested against a live server with `AUTH_CREDENTIALS` set.
- The prompt file is embedded in the binary. Rebuild the image to serve the new text.
- `npm run build` in `website` passed. Tests in `internal/handlers/mcp_handler` fail on this machine because it has no C compiler (`CGO_ENABLED=0`). The changes do not touch Go code.
