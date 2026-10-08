package mcphandler

import (
	"reflect"
	"strings"
	"testing"

	"github.com/DEEJ4Y/genkitkraft/internal/app/commands"
	"github.com/DEEJ4Y/genkitkraft/internal/app/queries"
)

// normalize makes snake_case json names and CamelCase Go names comparable.
func normalize(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "_", ""))
}

func fieldNames(t reflect.Type) map[string]bool {
	names := map[string]bool{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		names[normalize(f.Name)] = true
		if tag := strings.Split(f.Tag.Get("json"), ",")[0]; tag != "" {
			names[normalize(tag)] = true
		}
	}
	return names
}

// TestInputDTOsCoverAppParams guards against MCP tool inputs drifting from the
// app-layer params they feed. MCP DTOs are hand-written (no codegen), so a field
// added to a command's params must also be added to the MCP input, otherwise it
// cannot be set, and for replace-style updates it is silently wiped.
//
// allowMissing lists params fields that are intentionally not exposed.
func TestInputDTOsCoverAppParams(t *testing.T) {
	cases := []struct {
		name         string
		input        any
		params       any
		allowMissing []string
	}{
		{"agents_create", CreateAgentInput{}, commands.CreateAgentParams{}, nil},
		{"agents_update", UpdateAgentInput{}, commands.UpdateAgentParams{}, nil},
		{"agent_tools_update", UpdateAgentToolConfigInput{}, commands.UpdateAgentToolsParams{}, nil},
		{"http_tools_create", CreateHttpToolInput{}, commands.CreateHttpToolParams{}, nil},
		{"http_tools_update", UpdateHttpToolInput{}, commands.UpdateHttpToolParams{}, nil},
		{"providers_create", CreateProviderInput{}, commands.CreateProviderParams{}, nil},
		{"providers_update", UpdateProviderInput{}, commands.UpdateProviderParams{}, nil},
		{"mcp_servers_create", CreateMcpServerInput{}, commands.CreateMcpServerParams{}, nil},
		{"mcp_servers_update", UpdateMcpServerInput{}, commands.UpdateMcpServerParams{}, nil},
		{"prompts_create", CreatePromptInput{}, commands.CreatePromptParams{}, nil},
		{"prompts_update", UpdatePromptInput{}, commands.UpdatePromptParams{}, nil},
		{
			"playground_chat", PlaygroundChatInput{}, queries.ResolvePlaygroundConfigParams{},
			// ToolOverride is flattened into http_tool_ids / mcp_servers / built_in_tool_ids;
			// IncludeTools is always true for chat.
			[]string{"ToolOverride", "IncludeTools"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			have := fieldNames(reflect.TypeOf(tc.input))
			allowed := map[string]bool{}
			for _, a := range tc.allowMissing {
				allowed[normalize(a)] = true
			}
			pt := reflect.TypeOf(tc.params)
			for i := 0; i < pt.NumField(); i++ {
				name := pt.Field(i).Name
				if allowed[normalize(name)] {
					continue
				}
				if !have[normalize(name)] {
					t.Errorf("%s: MCP input %T has no field for app param %q; add it (or allowlist it if intentionally not exposed)",
						tc.name, tc.input, name)
				}
			}
		})
	}
}

func TestToResolvePlaygroundConfigParamsToolOverride(t *testing.T) {
	// No tool fields: no override, agent's saved tools are used.
	p := toResolvePlaygroundConfigParams(PlaygroundChatInput{AgentID: "a"})
	if p.ToolOverride != nil {
		t.Fatalf("expected nil ToolOverride when no tool fields are set, got %+v", p.ToolOverride)
	}
	if !p.IncludeTools {
		t.Fatal("expected IncludeTools to be true")
	}

	// Built-in tools alone trigger an override (e.g. enabling web_fetch).
	ids := []string{"web_fetch"}
	p = toResolvePlaygroundConfigParams(PlaygroundChatInput{AgentID: "a", BuiltInToolIDs: &ids})
	if p.ToolOverride == nil || len(p.ToolOverride.BuiltInToolIDs) != 1 || p.ToolOverride.BuiltInToolIDs[0] != "web_fetch" {
		t.Fatalf("expected web_fetch override, got %+v", p.ToolOverride)
	}

	// An explicit empty list is an override that disables the tools.
	empty := []string{}
	p = toResolvePlaygroundConfigParams(PlaygroundChatInput{AgentID: "a", BuiltInToolIDs: &empty})
	if p.ToolOverride == nil {
		t.Fatal("expected an empty built_in_tool_ids to still produce an override")
	}
}
