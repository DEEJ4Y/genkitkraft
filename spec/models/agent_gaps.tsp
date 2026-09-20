namespace Api.Models;

/** Category of a self-reported agent gap. */
enum GapCategory {
  Knowledge: "knowledge",
  Capability: "capability",
  Improvement: "improvement",
}

/** Review lifecycle status of a gap. */
enum GapStatus {
  Open: "open",
  Resolved: "resolved",
  Dismissed: "dismissed",
}

/** Reason category for dismissing a gap. */
enum GapDismissalCategory {
  Unrelated: "unrelated",
  InsufficientDetail: "insufficient_detail",
  Duplicate: "duplicate",
  Other: "other",
}

/** A conversation (and, when resolvable, message) a gap was observed in. */
model GapReference {
  /** ID of the playground/deploy session this gap was observed in, when one
   * exists. Absent for a report from the stateless deploy chat-completions
   * endpoint, which has no persisted conversation. */
  sessionId?: string;

  /** ID of the specific message, when it could be resolved. */
  messageId?: string;
}

/** A self-reported gap for an agent: a question it could not answer
 * reliably (knowledge), an action it could not perform (capability), or a
 * suggestion to automate more of the flow (improvement). */
model GapResponse {
  /** Unique gap ID. */
  id: string;

  /** ID of the agent this gap belongs to. */
  agentId: string;

  /** The kind of gap. */
  category: GapCategory;

  /** What the user asked or what task was attempted. */
  context: string;

  /** What was missing, blocked, or could be improved. */
  details: string;

  /** Optional suggestion for closing the gap. */
  suggestedResolution?: string;

  /** Review lifecycle status. */
  status: GapStatus;

  /** Reason category, present only when status is dismissed. */
  dismissalCategory?: GapDismissalCategory;

  /** Free-text dismissal reason, if provided. */
  dismissalReason?: string;

  /** Conversations this gap was observed in. */
  references: GapReference[];

  /** When this gap was first created. */
  createdAt: utcDateTime;

  /** When this gap was last updated. */
  updatedAt: utcDateTime;
}

/** Paginated list of gaps. */
model GapListResponse {
  /** Array of gaps. */
  gaps: GapResponse[];

  /** Total number of gaps for this agent. */
  total: int32;

  /** Number of gaps per page. */
  limit: int32;

  /** Number of gaps skipped. */
  offset: int32;
}

/** Request to update a gap's review status. Resolve, dismiss (with a
 * reason), or reopen (status: "open") — a gap dismissed with category
 * "unrelated" can never be reopened. */
model UpdateGapRequest {
  /** New status: "resolved", "dismissed", or "open" (to reopen). */
  status?: GapStatus;

  /** Required when status is "dismissed". */
  dismissalCategory?: GapDismissalCategory;

  /** Optional free-text reason, used together with dismissalCategory. */
  dismissalReason?: string;
}
