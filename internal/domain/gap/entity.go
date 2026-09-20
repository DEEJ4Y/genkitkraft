package gap

import "time"

// Category classifies the kind of blind spot an agent self-reported.
type Category string

const (
	// CategoryKnowledge means the agent could not answer reliably — it was
	// missing information or a source, not a tool or permission.
	CategoryKnowledge Category = "knowledge"
	// CategoryCapability means the agent could not perform a requested
	// action — it was missing a tool, permission, or integration.
	CategoryCapability Category = "capability"
	// CategoryImprovement means nothing failed — the agent is suggesting a
	// way to automate more of the flow.
	CategoryImprovement Category = "improvement"
)

// Status is the review lifecycle state of a Gap.
type Status string

const (
	StatusOpen      Status = "open"
	StatusResolved  Status = "resolved"
	StatusDismissed Status = "dismissed"
)

// Dismissal reason categories. DismissalUnrelated is terminal — a gap
// dismissed as unrelated can never be reopened, by a human or the
// background dedup pipeline.
const (
	DismissalUnrelated          = "unrelated"
	DismissalInsufficientDetail = "insufficient_detail"
	DismissalDuplicate          = "duplicate"
	DismissalOther              = "other"
)

// Gap is a self-reported blind spot an agent flagged during a conversation,
// deduplicated by the background pipeline into one row per distinct gap.
type Gap struct {
	ID                  string
	AgentID             string
	Category            Category
	Context             string
	Details             string
	SuggestedResolution string
	Status              Status
	DismissalCategory   string
	DismissalReason     string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// IsTerminal reports whether the gap can never be reopened again — true only
// for a gap dismissed as unrelated. Enforced centrally here so both the
// UI-triggered reopen and the background dedup pipeline's merge path agree.
func (g *Gap) IsTerminal() bool {
	return g.Status == StatusDismissed && g.DismissalCategory == DismissalUnrelated
}

// Reference ties a Gap to the conversation (and, when resolvable, message)
// it was observed in. SessionID is empty for a report from the stateless
// deploy chat-completions endpoint, which has no persisted conversation.
// MessageID is empty when the dedup pipeline could not resolve the
// in-flight message before the reference was written, or when SessionID
// itself is empty.
type Reference struct {
	ID        string
	GapID     string
	SessionID string
	MessageID string
	CreatedAt time.Time
}
