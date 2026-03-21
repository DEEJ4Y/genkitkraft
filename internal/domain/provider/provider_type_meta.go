package provider

// ConfigFieldInfo describes a single config field for a provider type.
type ConfigFieldInfo struct {
	Name        string
	Label       string
	Required    bool
	Placeholder string
	Sensitive   bool
}

// ProviderTypeInfo contains metadata about a provider type for UI rendering.
type ProviderTypeInfo struct {
	Type            ProviderType
	DisplayName     string
	RequiresAPIKey  bool
	RequiresBaseURL bool
	ConfigFields    []ConfigFieldInfo
	EnvVarHint      string
	ModelPrefix     string
	BaseURLDefault  string
	ComingSoon      bool
}

// ProviderTypeMeta is the static registry of all provider type metadata.
var ProviderTypeMeta = []ProviderTypeInfo{
	{
		Type:           GoogleAI,
		DisplayName:    "Google AI (Gemini)",
		RequiresAPIKey: true,
		EnvVarHint:     "GEMINI_API_KEY or GOOGLE_API_KEY",
		ModelPrefix:    "googleai",
	},
	{
		Type:        VertexAI,
		DisplayName: "Vertex AI",
		EnvVarHint:  "GOOGLE_CLOUD_PROJECT, GOOGLE_CLOUD_LOCATION",
		ModelPrefix: "vertexai",
		ConfigFields: []ConfigFieldInfo{
			{Name: "project", Label: "GCP Project ID", Required: true, Placeholder: "my-gcp-project"},
			{Name: "location", Label: "GCP Location", Required: true, Placeholder: "us-central1"},
		},
	},
	{
		Type:           OpenAI,
		DisplayName:    "OpenAI",
		RequiresAPIKey: true,
		EnvVarHint:     "OPENAI_API_KEY",
		ModelPrefix:    "openai",
	},
	{
		Type:           Anthropic,
		DisplayName:    "Anthropic",
		RequiresAPIKey: true,
		EnvVarHint:     "ANTHROPIC_API_KEY",
		ModelPrefix:    "anthropic",
	},
	{
		Type:            Ollama,
		DisplayName:     "Ollama",
		RequiresBaseURL: true,
		EnvVarHint:      "",
		ModelPrefix:     "ollama",
		BaseURLDefault:  "http://localhost:11434",
	},
	{
		Type:           XAI,
		DisplayName:    "xAI (Grok)",
		RequiresAPIKey: true,
		EnvVarHint:     "XAI_API_KEY",
		ModelPrefix:    "xai",
	},
	{
		Type:           DeepSeek,
		DisplayName:    "DeepSeek",
		RequiresAPIKey: true,
		EnvVarHint:     "DEEPSEEK_API_KEY",
		ModelPrefix:    "deepseek",
	},
	{
		Type:            AzureOpenAI,
		DisplayName:     "Azure OpenAI",
		RequiresAPIKey:  true,
		RequiresBaseURL: true,
		EnvVarHint:      "AZURE_OPEN_AI_API_KEY, AZURE_OPEN_AI_ENDPOINT",
		ModelPrefix:     "azure_openai",
		ConfigFields: []ConfigFieldInfo{
			{Name: "deployment_name", Label: "Deployment Name", Required: true, Placeholder: "gpt-4o"},
			{Name: "api_version", Label: "API Version", Required: true, Placeholder: "2024-10-21"},
		},
	},
	{
		Type:        Bedrock,
		DisplayName: "AWS Bedrock",
		EnvVarHint:  "AWS_REGION, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY",
		ModelPrefix: "bedrock",
		ConfigFields: []ConfigFieldInfo{
			{Name: "region", Label: "AWS Region", Required: true, Placeholder: "us-east-1"},
			{Name: "access_key_id", Label: "Access Key ID", Required: true, Sensitive: true},
			{Name: "secret_access_key", Label: "Secret Access Key", Required: true, Sensitive: true},
			{Name: "session_token", Label: "Session Token", Sensitive: true},
		},
	},
	{
		Type:            AzureAIFoundry,
		DisplayName:     "Azure AI Foundry",
		RequiresAPIKey:  true,
		RequiresBaseURL: true,
		EnvVarHint:      "AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_API_KEY",
		ModelPrefix:     "azureaifoundry",
	},
	{
		Type:            OpenAICompatible,
		DisplayName:     "OpenAI Compatible",
		RequiresAPIKey:  true,
		RequiresBaseURL: true,
		EnvVarHint:      "API key and base URL from your provider",
		ModelPrefix:     "openai",
	},
}
