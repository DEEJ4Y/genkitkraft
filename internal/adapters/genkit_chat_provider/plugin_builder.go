package genkitchatprovider

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/anthropic"
	"github.com/firebase/genkit/go/plugins/compat_oai"
	"github.com/firebase/genkit/go/plugins/compat_oai/openai"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/openai/openai-go/option"
	bedrock "github.com/xavidop/genkit-aws-bedrock-go"
	azureaifoundry "github.com/xavidop/genkit-azure-foundry-go"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
	chatprovider "github.com/DEEJ4Y/genkitkraft/internal/ports/chat_provider"
)

// pluginResult holds the genkit plugin and the model name to use for generation.
type pluginResult struct {
	plugin    api.Plugin
	modelName string
	// postInit is called after genkit.Init() for plugins that require explicit model registration.
	postInit func(g *genkit.Genkit) ai.Model
}

// buildPlugin maps a ChatRequest to the appropriate Genkit plugin and model name.
func buildPlugin(req chatprovider.ChatRequest) (*pluginResult, error) {
	pt := provider.ProviderType(req.ProviderType)

	switch pt {
	case provider.GoogleAI:
		return &pluginResult{
			plugin:    &googlegenai.GoogleAI{APIKey: req.APIKey},
			modelName: req.ModelID,
		}, nil

	case provider.VertexAI:
		cfg := parseVertexAIConfig(req.Config)
		return &pluginResult{
			plugin:    &googlegenai.VertexAI{ProjectID: cfg.Project, Location: cfg.Location},
			modelName: req.ModelID,
		}, nil

	case provider.OpenAI:
		return &pluginResult{
			plugin:    &openai.OpenAI{APIKey: req.APIKey},
			modelName: req.ModelID,
		}, nil

	case provider.Anthropic:
		return &pluginResult{
			plugin:    &anthropic.Anthropic{APIKey: req.APIKey},
			modelName: req.ModelID,
		}, nil

	case provider.Ollama:
		serverAddr := req.BaseURL
		if serverAddr == "" {
			serverAddr = "http://localhost:11434"
		}
		ollamaPlugin := &ollama.Ollama{ServerAddress: serverAddr}
		return &pluginResult{
			plugin:    ollamaPlugin,
			modelName: req.ModelID,
			postInit: func(g *genkit.Genkit) ai.Model {
				return ollamaPlugin.DefineModel(g, ollama.ModelDefinition{
					Name: req.ModelID,
					Type: "chat",
				}, &ai.ModelOptions{
					Supports: &ai.ModelSupports{
						Multiturn:  true,
						SystemRole: true,
					},
				})
			},
		}, nil

	case provider.XAI:
		baseURL := req.BaseURL
		if baseURL == "" {
			baseURL = "https://api.x.ai/v1"
		}
		return &pluginResult{
			plugin: &compat_oai.OpenAICompatible{
				Provider: "xai",
				APIKey:   req.APIKey,
				BaseURL:  baseURL,
			},
			modelName: req.ModelID,
		}, nil

	case provider.DeepSeek:
		baseURL := req.BaseURL
		if baseURL == "" {
			baseURL = "https://api.deepseek.com"
		}
		return &pluginResult{
			plugin: &compat_oai.OpenAICompatible{
				Provider: "deepseek",
				APIKey:   req.APIKey,
				BaseURL:  baseURL,
			},
			modelName: req.ModelID,
		}, nil

	case provider.AzureOpenAI:
		cfg := parseAzureOpenAIConfig(req.Config)
		// Azure OpenAI uses api-key header instead of Bearer token.
		// Build the deployment-based URL.
		apiVersion := cfg.APIVersion
		if apiVersion == "" {
			apiVersion = "2024-10-21"
		}
		deploymentName := cfg.DeploymentName
		if deploymentName == "" {
			deploymentName = req.ModelID
		}
		baseURL := fmt.Sprintf("%s/openai/deployments/%s", strings.TrimRight(req.BaseURL, "/"), deploymentName)

		return &pluginResult{
			plugin: &compat_oai.OpenAICompatible{
				Provider: "azure_openai",
				Opts: []option.RequestOption{
					option.WithHeader("api-key", req.APIKey),
					option.WithBaseURL(baseURL),
					option.WithHeader("api-version", apiVersion),
				},
			},
			modelName: req.ModelID,
		}, nil

	case provider.AzureAIFoundry:
		foundryPlugin := &azureaifoundry.AzureAIFoundry{
			Endpoint: req.BaseURL,
			APIKey:   req.APIKey,
		}
		return &pluginResult{
			plugin:    foundryPlugin,
			modelName: req.ModelID,
			postInit: func(g *genkit.Genkit) ai.Model {
				return foundryPlugin.DefineModel(g, azureaifoundry.ModelDefinition{
					Name: req.ModelID,
					Type: "chat",
				}, nil)
			},
		}, nil

	case provider.Bedrock:
		cfg := parseBedrockConfig(req.Config)
		awsCfg := aws.Config{
			Region:      cfg.Region,
			Credentials: credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.SessionToken),
		}
		bedrockPlugin := &bedrock.Bedrock{
			Region:    cfg.Region,
			AWSConfig: &awsCfg,
		}
		return &pluginResult{
			plugin:    bedrockPlugin,
			modelName: req.ModelID,
			postInit: func(g *genkit.Genkit) ai.Model {
				return bedrockPlugin.DefineModel(g, bedrock.ModelDefinition{
					Name: req.ModelID,
					Type: "chat",
				}, nil)
			},
		}, nil

	case provider.OpenAICompatible:
		opts := []option.RequestOption{}
		cfg := parseOpenAICompatibleConfig(req.Config)
		if cfg.Organization != "" {
			opts = append(opts, option.WithHeader("OpenAI-Organization", cfg.Organization))
		}
		if cfg.CustomHeaders != "" {
			for _, line := range strings.Split(cfg.CustomHeaders, "\n") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					opts = append(opts, option.WithHeader(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])))
				}
			}
		}
		return &pluginResult{
			plugin: &compat_oai.OpenAICompatible{
				Provider: "openai_compatible",
				APIKey:   req.APIKey,
				BaseURL:  req.BaseURL,
				Opts:     opts,
			},
			modelName: req.ModelID,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported provider type: %s", req.ProviderType)
	}
}

// --- Config parsers ---

type vertexAIConfig struct {
	Project  string `json:"project"`
	Location string `json:"location"`
}

func parseVertexAIConfig(raw json.RawMessage) vertexAIConfig {
	var cfg vertexAIConfig
	if len(raw) > 0 {
		json.Unmarshal(raw, &cfg)
	}
	return cfg
}

type azureOpenAIConfig struct {
	DeploymentName string `json:"deployment_name"`
	APIVersion     string `json:"api_version"`
}

func parseAzureOpenAIConfig(raw json.RawMessage) azureOpenAIConfig {
	var cfg azureOpenAIConfig
	if len(raw) > 0 {
		json.Unmarshal(raw, &cfg)
	}
	return cfg
}

type bedrockConfig struct {
	Region          string `json:"region"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey  string `json:"secret_access_key"`
	SessionToken    string `json:"session_token"`
}

func parseBedrockConfig(raw json.RawMessage) bedrockConfig {
	var cfg bedrockConfig
	if len(raw) > 0 {
		json.Unmarshal(raw, &cfg)
	}
	return cfg
}

type openAICompatibleConfig struct {
	Organization  string `json:"organization"`
	CustomHeaders string `json:"custom_headers"`
}

func parseOpenAICompatibleConfig(raw json.RawMessage) openAICompatibleConfig {
	var cfg openAICompatibleConfig
	if len(raw) > 0 {
		json.Unmarshal(raw, &cfg)
	}
	return cfg
}
