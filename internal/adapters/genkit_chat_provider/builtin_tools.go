package genkitchatprovider

import (
	"fmt"
	"io"
	"net/http"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/firebase/genkit/go/ai"
)

// buildBuiltInTools creates Genkit tool references for the specified built-in tool IDs.
func buildBuiltInTools(ids []string) []ai.ToolRef {
	var tools []ai.ToolRef
	for _, id := range ids {
		switch id {
		case "web_fetch":
			tools = append(tools, buildWebFetchTool())
		}
	}
	return tools
}

func buildWebFetchTool() ai.Tool {
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
		return fetchAsMarkdown(url)
	}

	return ai.NewTool(name, description, toolFn, ai.WithInputSchema(inputSchema))
}

func fetchAsMarkdown(rawURL string) (string, error) {
	encodedURL, err := sanitizeURL(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	resp, err := http.Get(encodedURL)
	if err != nil {
		return "", fmt.Errorf("fetching URL: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading body: %w", err)
	}

	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(string(body))
	if err != nil {
		return "", fmt.Errorf("converting to markdown: %w", err)
	}

	return markdown, nil
}
