package genkitchatprovider

import "testing"

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "already valid URL unchanged",
			input: "https://example.com/path?q=hello",
			want:  "https://example.com/path?q=hello",
		},
		{
			name:  "spaces in path are encoded",
			input: "https://example.com/path with spaces/page",
			want:  "https://example.com/path%20with%20spaces/page",
		},
		{
			name:  "spaces in query value are encoded",
			input: "https://example.com/search?q=hello world",
			want:  "https://example.com/search?q=hello+world",
		},
		{
			name:  "existing percent encoding not double-encoded",
			input: "https://example.com/path%20already/page",
			want:  "https://example.com/path%20already/page",
		},
		{
			name:  "unicode in path is encoded",
			input: "https://example.com/café",
			want:  "https://example.com/caf%C3%A9",
		},
		{
			name:  "unicode in query is encoded",
			input: "https://example.com/search?q=café",
			want:  "https://example.com/search?q=caf%C3%A9",
		},
		{
			name:  "multiple query params preserved",
			input: "https://example.com/api?name=John Doe&age=30",
			want:  "https://example.com/api?age=30&name=John+Doe",
		},
		{
			name:  "empty query string left empty",
			input: "https://example.com/path",
			want:  "https://example.com/path",
		},
		{
			name:  "fragment preserved",
			input: "https://example.com/page#section one",
			want:  "https://example.com/page#section%20one",
		},
		{
			name:    "invalid URL returns error",
			input:   "://missing-scheme",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitizeURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("sanitizeURL(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("sanitizeURL(%q)\n  got  = %q\n  want = %q", tt.input, got, tt.want)
			}
		})
	}
}
