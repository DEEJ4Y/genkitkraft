package config

import (
	"os"
	"strings"
)

type Config struct {
	Server     ServerConfig
	Auth       AuthConfig
	Database   DatabaseConfig
	Cache      CacheConfig
	Encryption EncryptionConfig
	Deploy     DeployConfig
}

type DatabaseConfig struct {
	Provider string // "sqlite" (default), "postgres", "mysql", "mariadb"
	Path     string // SQLite only
	URL      string // required for non-SQLite providers
}

// CacheConfig selects the backing store for sessions and login rate limiting.
// The default is process-local; multi-instance deployments need a shared provider
// so state is visible to every instance. Web-fetch results are deliberately not
// covered by this setting — they stay process-local so agent tool traffic cannot
// exhaust the cache that authentication depends on.
type CacheConfig struct {
	Provider string // "memory" (default), "redis", "valkey"
	URL      string // required for non-memory providers
}

type EncryptionConfig struct {
	Key string
}

type ServerConfig struct {
	Port string
}

type DeployConfig struct {
	APIKeys map[string]struct{}
}

type AuthCredential struct {
	Username string
	Password string
}

type AuthConfig struct {
	Credentials []AuthCredential
}

func Load() Config {
	return Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Auth: loadAuthConfig(),
		Database: DatabaseConfig{
			Provider: getEnv("DATABASE_PROVIDER", "sqlite"),
			Path:     getEnv("DATABASE_PATH", "/data/app.db"),
			URL:      os.Getenv("DATABASE_URL"),
		},
		Cache: CacheConfig{
			Provider: getEnv("CACHE_PROVIDER", "memory"),
			URL:      os.Getenv("CACHE_URL"),
		},
		Encryption: EncryptionConfig{
			Key: os.Getenv("ENCRYPTION_KEY"),
		},
		Deploy: loadDeployConfig(),
	}
}

func loadAuthConfig() AuthConfig {
	raw := os.Getenv("AUTH_CREDENTIALS")
	if raw == "" {
		return AuthConfig{}
	}

	pairs := strings.Split(raw, ",")
	creds := make([]AuthCredential, 0, len(pairs))

	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) != 2 {
			continue
		}
		username := strings.TrimSpace(parts[0])
		password := strings.TrimSpace(parts[1])
		if username == "" || password == "" {
			continue
		}
		creds = append(creds, AuthCredential{
			Username: username,
			Password: password,
		})
	}

	return AuthConfig{Credentials: creds}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadDeployConfig() DeployConfig {
	raw := os.Getenv("PUBLIC_API_KEY")
	keys := make(map[string]struct{})
	if raw == "" {
		return DeployConfig{APIKeys: keys}
	}
	for _, k := range strings.Split(raw, ",") {
		k = strings.TrimSpace(k)
		if k != "" {
			keys[k] = struct{}{}
		}
	}
	return DeployConfig{APIKeys: keys}
}
