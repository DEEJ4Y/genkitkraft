package builtintool

// BuiltInTool represents a tool whose functionality is built into GenKitKraft.
type BuiltInTool struct {
	ID          string
	Name        string
	Description string
}

// Registry contains all available built-in tools.
var Registry = []BuiltInTool{
	{
		ID:   "web_fetch",
		Name: "Web Fetch",
		Description: "Fetches static content from a URL and returns it as Markdown. " +
			"Use this tool to browse documentation, reference pages, and other web resources " +
			"to provide context-rich and accurate answers. " +
			"Note: This tool only supports static HTML content. It does not execute JavaScript " +
			"or render dynamic content (no headless browser). Pages that require client-side " +
			"rendering will return incomplete results.",
	},
}

// GetByID returns a built-in tool by its ID, or nil if not found.
func GetByID(id string) *BuiltInTool {
	for i := range Registry {
		if Registry[i].ID == id {
			return &Registry[i]
		}
	}
	return nil
}

// IsValid returns true if the given ID corresponds to a known built-in tool.
func IsValid(id string) bool {
	return GetByID(id) != nil
}
