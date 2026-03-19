package config

import (
	"os"
	"strings"
)

type Config struct {
	Server   ServerConfig
	Auth     AuthConfig
	Database DatabaseConfig
}

type DatabaseConfig struct {
	Path string
}

type ServerConfig struct {
	Port string
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
			Path: getEnv("DATABASE_PATH", "/data/app.db"),
		},
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
