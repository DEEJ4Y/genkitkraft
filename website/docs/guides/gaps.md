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

## Reviewing gaps

Open the agent's **Gaps** tab to see every reported gap, with its category, context, details, and any suggested resolution. Each gap also lists the playground or deploy sessions it was observed in — reports from the stateless Deploy chat-completions endpoint show up as "(stateless call)" instead, since there's no conversation to point to.

From here you can:

- **Resolve** a gap once you've addressed it (e.g. added the missing source or tool).
- **Dismiss** a gap with a reason — unrelated, insufficient detail, duplicate, or other. A gap dismissed as **unrelated** is permanent and can't be reopened; other dismissals can be reopened later.
- **Reopen** a resolved or dismissed (non-unrelated) gap.

## Scope and limitations

- Gap reporting is available on the Playground, stateful Deploy session APIs, and the stateless Deploy chat-completions endpoint. The stateless endpoint has no persisted conversation, so reports from it carry no session or message reference — the gap's own category/context/details still capture everything needed to review it.
- Gaps are scoped to the reporting agent — they aren't shared or correlated across agents.
- Two near-simultaneous reports for the same underlying gap can occasionally create separate entries instead of merging; review the Gaps tab periodically to catch duplicates.
