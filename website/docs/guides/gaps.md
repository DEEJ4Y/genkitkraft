---
sidebar_position: 7
---

# Gaps

Agents can self-report gaps they notice during a conversation — without changing the answer the end user actually receives. A **gap** is one of three things:

- **Knowledge** — the agent couldn't answer a question reliably (missing information or a source).
- **Capability** — the agent couldn't perform a requested action (missing tool, permission, or integration).
- **Improvement** — nothing failed, but the agent sees an opportunity to automate more of the flow.

Gaps are reviewed on the agent's **Gaps** tab, so you can find missing sources, missing tools, and automation ideas without waiting for an end user to complain.

## Enabling Gap Reporting

1. Open an agent for editing.
2. On the **Configuration** tab, toggle **Enable Gap Reporting**.
3. Click **Save Changes**.

Once enabled, the agent gets a `report_gap` tool it can call during a conversation. Calling this tool never blocks or alters the agent's answer to the user — it fires in the background, and generation continues normally.

Enabling the flag also appends an internal instruction block to the end of the agent's system prompt, describing the three categories and when to call `report_gap` — including reporting a capability gap even when the agent's own instructions already document the limitation, and reporting every occurrence rather than assuming it was already reported elsewhere. These instructions are never shown to end users.

## How gaps are deduplicated

Reported gaps aren't written straight to the list. A background pipeline reviews each new report against the agent's existing open gaps and either merges it into a matching gap (combining details) or creates a new one — so repeatedly hitting the same blind spot doesn't flood the tab with duplicates. This runs asynchronously, using the agent's own configured provider and model; it never adds latency to the conversation the gap was reported from.

The server runs one review at a time for each agent. If one reply reports the same gap twice, the second review sees the gap that the first review created, so both reports merge into one gap.

## Reviewing gaps

Open the agent's **Gaps** tab to see every reported gap, with its category, context, details, and any suggested resolution. Each gap also lists the playground or deploy sessions it was observed in — reports from the stateless Deploy chat-completions endpoint show up as "(stateless call)" instead, since there's no conversation to point to.

From here you can:

- **Resolve** a gap once you've addressed it (e.g. added the missing source or tool).
- **Dismiss** a gap with a reason — unrelated, insufficient detail, duplicate, or other. A gap dismissed as **unrelated** is permanent and can't be reopened; other dismissals can be reopened later.
- **Reopen** a resolved or dismissed (non-unrelated) gap.

You can do the same from an MCP client with the `gaps_list`, `gaps_get`, `gaps_resolve`, `gaps_dismiss`, and `gaps_reopen` tools. See the [MCP Quickstart](/docs/guides/mcp-quickstart#gaps).

### When a gap comes back

A new report that matches a resolved or dismissed gap reopens it. The gap keeps its triage history:

- The tab shows **Reopened after resolved** or **Reopened after dismissed**.
- The earlier dismissal category and reason stay visible as "Previously dismissed".
- The tab shows when the gap was last reopened and when it was last reported.

A gap that keeps coming back after you resolved it may need a better fix. A gap dismissed as **unrelated** never reopens. A new report that matches it creates a new gap.

## Scope and limitations

- Gap reporting is available on the Playground, stateful Deploy session APIs, and the stateless Deploy chat-completions endpoint. The stateless endpoint has no persisted conversation, so reports from it carry no session or message reference — the gap's own category/context/details still capture everything needed to review it.
- Gaps are scoped to the reporting agent — they aren't shared or correlated across agents.
- The one-review-at-a-time rule works inside one server process. If you run more than one server instance on the same database, near-simultaneous reports can still create separate entries. Review the Gaps tab to catch duplicates.
