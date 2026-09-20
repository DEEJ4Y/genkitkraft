package genkitchatprovider

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/firebase/genkit/go/ai"

	gapreporter "github.com/DEEJ4Y/genkitkraft/internal/ports/gap_reporter"
)

// buildBuiltInTools creates Genkit tool references for the specified built-in
// tool IDs, plus the report_gap tool when gap reporting is enabled for this
// conversation. report_gap is deliberately not part of ids/the built-in tool
// registry — it's gated by a per-agent feature flag, not a Tools-tab
// assignment — and is only offered when a session exists to attach it to.
func (cp *ChatProvider) buildBuiltInTools(ids []string, sessionID string, gapReportingEnabled bool) []ai.ToolRef {
	var tools []ai.ToolRef
	for _, id := range ids {
		switch id {
		case "web_fetch":
			tools = append(tools, cp.buildWebFetchTool())
		}
	}
	if gapReportingEnabled && sessionID != "" && cp.gapReporter != nil {
		tools = append(tools, cp.buildReportGapTool(sessionID))
	}
	return tools
}

func (cp *ChatProvider) buildReportGapTool(sessionID string) ai.Tool {
	name := "report_gap"
	description := "Report a gap you noticed in this conversation: a question you could not answer " +
		"reliably (category 'knowledge'), an action you were asked to perform but could not " +
		"(category 'capability' — e.g. a missing tool, permission, or integration), or an idea for " +
		"automating more of this flow with no failure involved (category 'improvement'). This does " +
		"not change your answer to the user — answer them as best you can regardless, then call this " +
		"tool to flag the gap for review. Calling this tool has no visible effect on the conversation."

	inputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"category": map[string]any{
				"type":        "string",
				"enum":        []any{"knowledge", "capability", "improvement"},
				"description": "The kind of gap: 'knowledge' (couldn't answer reliably), 'capability' (couldn't perform an action), or 'improvement' (a suggestion, no failure occurred).",
			},
			"context": map[string]any{
				"type":        "string",
				"description": "What the user asked or what task you were attempting.",
			},
			"details": map[string]any{
				"type":        "string",
				"description": "What was missing, blocked, or could be improved.",
			},
			"suggested_resolution": map[string]any{
				"type":        "string",
				"description": "Optional: your suggestion for closing the gap (e.g. a source to add, a tool to grant, a step to automate).",
			},
		},
		"required": []any{"category", "context", "details"},
	}

	toolFn := func(toolCtx *ai.ToolContext, args any) (any, error) {
		argsMap, ok := args.(map[string]any)
		if !ok {
			return "Gap not recorded: invalid arguments.", nil
		}
		category, _ := argsMap["category"].(string)
		context_, _ := argsMap["context"].(string)
		details, _ := argsMap["details"].(string)
		if category == "" || context_ == "" || details == "" {
			return "Gap not recorded: category, context, and details are required.", nil
		}
		suggestedResolution, _ := argsMap["suggested_resolution"].(string)

		cp.reportGap(toolCtx.Context, sessionID, category, context_, details, suggestedResolution)
		return "Gap recorded for review.", nil
	}

	return ai.NewTool(name, description, toolFn, ai.WithInputSchema(inputSchema))
}

// reportGap is split out of the tool closure so it's reachable from a test
// without standing up Genkit's action machinery. It never fails the tool
// call — a broken reporter must not disrupt the live response — so any
// error is only logged by the underlying Reporter implementation.
func (cp *ChatProvider) reportGap(ctx context.Context, sessionID, category, context_, details, suggestedResolution string) {
	_ = cp.gapReporter.Report(ctx, gapreporter.ReportParams{
		SessionID:           sessionID,
		Category:            category,
		Context:             context_,
		Details:             details,
		SuggestedResolution: suggestedResolution,
	})
}

func (cp *ChatProvider) buildWebFetchTool() ai.Tool {
	name := "web_fetch"
	description := "Fetches static content from a URL and returns it as Markdown. " +
		"Note: This tool only supports static HTML content. It does not execute JavaScript " +
		"or render dynamic content (no headless browser). Pages that require client-side " +
		"rendering will return incomplete results."

	inputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url": map[string]any{
				"type":        "string",
				"description": "The URL to fetch content from.",
			},
		},
		"required": []any{"url"},
	}

	toolFn := func(toolCtx *ai.ToolContext, args any) (any, error) {
		argsMap, ok := args.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid arguments: expected object")
		}
		url, ok := argsMap["url"].(string)
		if !ok || url == "" {
			return nil, fmt.Errorf("invalid arguments: 'url' is required and must be a string")
		}

		return cp.webFetch(toolCtx.Context, url)
	}

	return ai.NewTool(name, description, toolFn, ai.WithInputSchema(inputSchema))
}

// webFetch is split out of the tool closure so the cache path is reachable from a
// test without standing up Genkit's action machinery.
//
// A cache error is treated as a miss and the page re-fetched. This cache is a
// latency optimisation, not a source of truth, so an outage should cost a round
// trip rather than fail the tool call — unlike session lookup, where a miss and an
// outage mean genuinely different things.
func (cp *ChatProvider) webFetch(ctx context.Context, url string) (string, error) {
	if cached, ok, err := cp.cache.Get(ctx, url); err == nil && ok {
		return cached, nil
	}

	result, err := fetchAsMarkdown(url)
	if err != nil {
		return "", err
	}

	_ = cp.cache.Set(ctx, url, result, time.Hour)
	return result, nil
}

func fetchAsMarkdown(rawURL string) (string, error) {
	encodedURL, err := sanitizeURL(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	resp, err := httpClient.Get(encodedURL)
	if err != nil {
		return "", fmt.Errorf("fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Sprintf("Fetch failed with status %d", resp.StatusCode), nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return "", fmt.Errorf("reading body: %w", err)
	}
	truncated := len(body) == maxResponseBytes

	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(string(body))
	if err != nil {
		return "", fmt.Errorf("converting to markdown: %w", err)
	}

	if strings.TrimSpace(markdown) == "" {
		return "No content received. Page may be dynamically rendered.", nil
	}

	if truncated {
		// Say so explicitly: this text goes to a model that will summarize it, and a
		// silently truncated article reads as a complete one.
		markdown += "\n\n[Content truncated: the page exceeded the 1MB fetch limit.]"
	}

	return markdown, nil
}
