package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/gap"
	agentrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_repo"
	chatprovider "github.com/DEEJ4Y/genkitkraft/internal/ports/chat_provider"
	"github.com/DEEJ4Y/genkitkraft/internal/ports/encryptor"
	gaprepo "github.com/DEEJ4Y/genkitkraft/internal/ports/gap_repo"
	playgroundrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/playground_repo"
	providerrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/provider_repo"
)

// gapMessageResolveRetries/Delay bound how long the dedup pipeline waits for
// the in-flight assistant message to be persisted before falling back to a
// session-only reference (message_id left empty). Only matters on the
// non-streaming chat path, where the message row is created after
// generation completes rather than before.
const (
	gapMessageResolveRetries = 3
	gapMessageResolveDelay   = 200 * time.Millisecond
)

// gapDedupSystemPrompt instructs the model to decide whether a new report
// duplicates an existing open gap or is genuinely new. No vector DB or
// embeddings are involved — this reuses the reporting agent's own model to
// reason over a small, agent-scoped list of existing gaps instead.
const gapDedupSystemPrompt = `You deduplicate self-reported gaps for an AI agent. You will be given a new gap report and a list of the agent's existing gaps. Decide whether the new report is the same underlying gap as one already listed (a "merge") or genuinely new (a "create").

Respond with ONLY a single JSON object, no markdown fences, no commentary:
{"action": "merge" or "create", "gap_id": "<id of the existing gap, only if action is merge>", "category": "knowledge" or "capability" or "improvement", "context": "<merged or original context>", "details": "<merged or original details, combining new information if merging>", "suggested_resolution": "<merged or original suggestion, may be empty>"}

Merge only when the new report is clearly about the same underlying gap (same missing information, same missing capability, or the same improvement idea) as an existing one. When merging, combine details so nothing from either report is lost. When uncertain, prefer "create".`

// agentLocks serializes dedup passes per agent. Each pass reads the agent's
// gaps, asks the LLM for a decision, then writes. Without the lock, two
// concurrent passes (for example two report_gap calls in one turn) both read
// before either writes, so a duplicate report creates a second gap.
// The lock lives in this process only. It does not protect a deployment with
// more than one server instance on a shared database.
type agentLocks struct {
	mu      sync.Mutex
	entries map[string]*agentLockEntry
}

type agentLockEntry struct {
	sem  chan struct{}
	refs int
}

// lock waits for the agent's lock or for ctx to end. On success it returns
// the unlock function.
func (l *agentLocks) lock(ctx context.Context, agentID string) (func(), error) {
	l.mu.Lock()
	if l.entries == nil {
		l.entries = make(map[string]*agentLockEntry)
	}
	e, ok := l.entries[agentID]
	if !ok {
		e = &agentLockEntry{sem: make(chan struct{}, 1)}
		l.entries[agentID] = e
	}
	e.refs++
	l.mu.Unlock()

	release := func() {
		l.mu.Lock()
		e.refs--
		if e.refs == 0 {
			delete(l.entries, agentID)
		}
		l.mu.Unlock()
	}

	select {
	case e.sem <- struct{}{}:
		return func() {
			<-e.sem
			release()
		}, nil
	case <-ctx.Done():
		release()
		return nil, ctx.Err()
	}
}

type RunGapDedupParams struct {
	SessionID           string
	AgentID             string
	Category            string
	Context             string
	Details             string
	SuggestedResolution string
}

type gapDedupDecision struct {
	Action              string `json:"action"`
	GapID               string `json:"gap_id"`
	Category            string `json:"category"`
	Context             string `json:"context"`
	Details             string `json:"details"`
	SuggestedResolution string `json:"suggested_resolution"`
}

// RunGapDedupCommand is the only thing that ever writes to the gaps tables.
// It runs as a detached background pipeline (see ReportGapCommand) using the
// reporting agent's own provider/model/credentials — never through the
// live conversation's tool-calling loop, and never seen by the live agent.
type RunGapDedupCommand struct {
	gapRepo        gaprepo.GapRepository
	agentRepo      agentrepo.AgentRepository
	providerRepo   providerrepo.ProviderRepository
	enc            encryptor.Encryptor
	playgroundRepo playgroundrepo.PlaygroundRepository
	chatProvider   chatprovider.ChatProvider
	logger         zerolog.Logger
	locks          agentLocks
}

func NewRunGapDedupCommand(
	gapRepo gaprepo.GapRepository,
	agentRepo agentrepo.AgentRepository,
	providerRepo providerrepo.ProviderRepository,
	enc encryptor.Encryptor,
	playgroundRepo playgroundrepo.PlaygroundRepository,
	chatProvider chatprovider.ChatProvider,
	logger zerolog.Logger,
) *RunGapDedupCommand {
	return &RunGapDedupCommand{
		gapRepo:        gapRepo,
		agentRepo:      agentRepo,
		providerRepo:   providerRepo,
		enc:            enc,
		playgroundRepo: playgroundRepo,
		chatProvider:   chatProvider,
		logger:         logger,
	}
}

// Execute runs one dedup pass. ctx bounds the wait for the per-agent lock.
// GapDedupTimeout bounds the pass itself and starts after the lock is taken.
func (c *RunGapDedupCommand) Execute(ctx context.Context, params RunGapDedupParams) error {
	unlock, err := c.locks.lock(ctx, params.AgentID)
	if err != nil {
		return fmt.Errorf("waiting for earlier dedup passes of the agent: %w", err)
	}
	defer unlock()

	ctx, cancel := context.WithTimeout(ctx, GapDedupTimeout)
	defer cancel()

	a, err := c.agentRepo.GetByID(ctx, params.AgentID)
	if err != nil {
		return fmt.Errorf("loading agent: %w", err)
	}

	p, err := c.providerRepo.GetByID(ctx, a.ProviderID)
	if err != nil {
		return fmt.Errorf("loading provider: %w", err)
	}

	apiKey := ""
	if p.APIKey != nil {
		apiKey, err = c.enc.Decrypt(*p.APIKey)
		if err != nil {
			return fmt.Errorf("decrypting api key: %w", err)
		}
	}

	existing, err := c.gapRepo.List(ctx, params.AgentID, "", 50, 0)
	if err != nil {
		return fmt.Errorf("listing existing gaps: %w", err)
	}

	resp, err := c.chatProvider.Chat(ctx, chatprovider.ChatRequest{
		ProviderType: string(p.ProviderType),
		APIKey:       apiKey,
		BaseURL:      p.BaseURL,
		Config:       p.RawConfig,
		ModelID:      a.ModelID,
		SystemPrompt: gapDedupSystemPrompt,
		Messages: []chatprovider.ChatMessage{
			{Role: "user", Content: buildGapDedupPrompt(existing, params)},
		},
	})
	if err != nil {
		return fmt.Errorf("dedup generation: %w", err)
	}

	decision, err := parseGapDedupDecision(resp)
	if err != nil {
		c.logger.Warn().Err(err).Str("agent_id", params.AgentID).Msg("gap dedup: could not parse model decision, creating new gap")
		decision = gapDedupDecision{
			Action:              "create",
			Category:            params.Category,
			Context:             params.Context,
			Details:             params.Details,
			SuggestedResolution: params.SuggestedResolution,
		}
	}

	gapID, err := c.applyDedupDecision(ctx, params, decision)
	if err != nil {
		return fmt.Errorf("applying dedup decision: %w", err)
	}

	ref := &gap.Reference{
		GapID:     gapID,
		SessionID: params.SessionID,
	}
	if params.SessionID != "" {
		ref.MessageID = c.resolveMessageID(ctx, params.SessionID)
	}
	if err := c.gapRepo.AddReference(ctx, ref); err != nil {
		return fmt.Errorf("adding gap reference: %w", err)
	}

	return nil
}

// applyDedupDecision merges into the target gap named by decision.GapID, or
// creates a new gap — falling through to create when the named gap doesn't
// exist, belongs to a different agent (the model hallucinated an id), or is
// terminal (dismissed as unrelated, which can never be reopened).
func (c *RunGapDedupCommand) applyDedupDecision(ctx context.Context, params RunGapDedupParams, decision gapDedupDecision) (string, error) {
	if decision.Action == "merge" && decision.GapID != "" {
		g, err := c.gapRepo.GetByID(ctx, decision.GapID)
		if err == nil && g.AgentID == params.AgentID && !g.IsTerminal() {
			g.Context = firstNonEmpty(decision.Context, g.Context)
			g.Details = firstNonEmpty(decision.Details, g.Details)
			if decision.SuggestedResolution != "" {
				g.SuggestedResolution = decision.SuggestedResolution
			}
			// Keep the dismissal category and reason as triage history. Reopen
			// records which status the gap came from.
			now := time.Now().UTC()
			g.Reopen(now)
			g.LastReportedAt = now
			if err := c.gapRepo.Update(ctx, g); err != nil {
				return "", err
			}
			return g.ID, nil
		}
	}

	category := gap.Category(decision.Category)
	if category == "" {
		category = gap.Category(params.Category)
	}
	g := &gap.Gap{
		AgentID:             params.AgentID,
		Category:            category,
		Context:             firstNonEmpty(decision.Context, params.Context),
		Details:             firstNonEmpty(decision.Details, params.Details),
		SuggestedResolution: firstNonEmpty(decision.SuggestedResolution, params.SuggestedResolution),
	}
	if err := c.gapRepo.Create(ctx, g); err != nil {
		return "", err
	}
	return g.ID, nil
}

func (c *RunGapDedupCommand) resolveMessageID(ctx context.Context, sessionID string) string {
	for i := 0; i < gapMessageResolveRetries; i++ {
		msg, err := c.playgroundRepo.GetLatestMessageBySession(ctx, sessionID)
		if err == nil {
			return msg.ID
		}
		time.Sleep(gapMessageResolveDelay)
	}
	return ""
}

func buildGapDedupPrompt(existing []*gap.Gap, params RunGapDedupParams) string {
	var b strings.Builder
	b.WriteString("New gap report:\n")
	fmt.Fprintf(&b, "category: %s\ncontext: %s\ndetails: %s\nsuggested_resolution: %s\n\n",
		params.Category, params.Context, params.Details, params.SuggestedResolution)

	if len(existing) == 0 {
		b.WriteString("Existing gaps: none.\n")
		return b.String()
	}

	b.WriteString("Existing gaps:\n")
	for _, g := range existing {
		fmt.Fprintf(&b, "- id: %s | status: %s | category: %s | context: %s | details: %s\n",
			g.ID, g.Status, g.Category, g.Context, g.Details)
	}
	return b.String()
}

func parseGapDedupDecision(resp string) (gapDedupDecision, error) {
	start := strings.IndexByte(resp, '{')
	end := strings.LastIndexByte(resp, '}')
	if start == -1 || end == -1 || end < start {
		return gapDedupDecision{}, fmt.Errorf("no JSON object found in model response")
	}

	var decision gapDedupDecision
	if err := json.Unmarshal([]byte(resp[start:end+1]), &decision); err != nil {
		return gapDedupDecision{}, fmt.Errorf("unmarshaling decision: %w", err)
	}
	if decision.Context == "" || decision.Details == "" {
		return gapDedupDecision{}, fmt.Errorf("decision missing required fields")
	}
	return decision, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
