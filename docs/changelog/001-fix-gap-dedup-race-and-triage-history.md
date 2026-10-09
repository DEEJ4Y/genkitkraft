# Fix gap dedup race and lost triage history

**Date:** 2026-10-09
**Branch:** fix/50-51-gap-dedup-race-and-triage-history

## Summary

Concurrent duplicate `report_gap` calls now create one gap (issue #50). A new report on a resolved or dismissed gap now keeps the triage history (issue #51).

## Changes

- `internal/app/commands/run_gap_dedup.go`: Add a lock for each agent. Only one dedup pass runs for an agent at a time. The merge path now calls `Gap.Reopen` and sets `LastReportedAt`. It no longer clears the dismissal fields.
- `internal/app/commands/report_gap.go`: Change `GapDedupTimeout` from 2 minutes to 5 minutes. The timeout starts after the pass takes the lock. Add `GapDedupLockWait` (10 minutes) for the wait in the queue.
- `internal/app/commands/reopen_gap.go`: Call `Gap.Reopen`. Fix the comment about the dedup path.
- `internal/domain/gap/entity.go`: Add `ReopenedFrom`, `ReopenedAt`, `LastReportedAt`, and the `Reopen` method.
- `internal/adapters/sqlite_db/migrations/014_create_agent_gaps.sql`, `internal/adapters/postgres_db/migrations/014_create_agent_gaps.sql`, `internal/adapters/mysql_db/migrations/014_create_agent_gaps.sql`: Add `reopened_from`, `reopened_at`, and `last_reported_at` to `agent_gaps`. The migration is edited in place.
- `internal/adapters/sqlite_gap/repository.go`, `internal/adapters/postgres_gap/repository.go`, `internal/adapters/mysql_gap/repository.go`: Read and write the new columns.
- `spec/models/agent_gaps.tsp`, `spec/tsp-output/schema/openapi.yaml`, `internal/api/gen/types.gen.go`, `ui/lib/api/schema.d.ts`: Add `reopenedFrom`, `reopenedAt`, and `lastReportedAt` to `GapResponse`.
- `internal/handlers/http_handler/type_conversion.go`: Map the new fields.
- `internal/handlers/mcp_handler/gap_tools.go`: Add the new fields to `GapOutput`. Update the `gaps_reopen` description.
- `ui/components/AgentGapsTab.tsx`: Show a "Reopened after ..." badge, the previous dismissal, and the last reported time.
- `website/docs/guides/gaps.md`: Describe the one-review-at-a-time rule and the reopen history.
- `internal/app/commands/run_gap_dedup_test.go`, `internal/app/commands/reopen_gap_test.go`, `internal/domain/gap/entity_test.go`, `internal/adapters/postgres_gap/repository_test.go`, `internal/adapters/mysql_gap/repository_test.go`: Add and update tests.

## Notes

- Edit in place of migration 014: a database that already ran the old 014 must be reset.
- The lock works in one server process only. Two server instances on one database can still create duplicates.
- A manual reopen from the UI also keeps the dismissal history.
- All Go tests passed in a `golang:1.26` Docker container with CGO on, including the SQLite tests. The race detector (`-race`) passed on `internal/app/commands` and `internal/domain/gap`. The Postgres and MySQL integration tests (`-tags integration`) and `npm run build` in `ui/` also passed.
- A manual test through MCP on a MySQL docker compose stack passed 7 of 7 tests (race, control, reopen after dismissal, reopen after resolution, final gap, manual reopen, open gap). Not tested: the UI in a browser, and two server instances on one database.
- Not done: batching reports for each turn, and a uniqueness constraint.
