package genkitchatprovider

import (
	"fmt"
	"io"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/firebase/genkit/go/ai"
)

// buildBuiltInTools creates Genkit tool references for the specified built-in tool IDs.
func (cp *ChatProvider) buildBuiltInTools(ids []string) []ai.ToolRef {
	var tools []ai.ToolRef
	for _, id := range ids {
		switch id {
		case "web_fetch":
			tools = append(tools, cp.buildWebFetchTool())
		}
	}
	return tools
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

		if cached, ok := cp.cache.Get(toolCtx.Context, url); ok {
			return cached, nil
		}

		result, err := fetchAsMarkdown(url)
		if err != nil {
			return nil, err
		}

		_ = cp.cache.Set(toolCtx.Context, url, result, time.Hour)
		return result, nil
	}

	return ai.NewTool(name, description, toolFn, ai.WithInputSchema(inputSchema))
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
