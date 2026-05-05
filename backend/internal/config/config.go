package config

import (
	"os"
	"strings"
)

type Config struct {
	Environment      string
	Address          string
	DatabaseURL      string
	CORSOrigins      []string
	JWTSecret        string
	FrontendURL      string
	SidecarURL       string
	StorageEndpoint  string
	StorageAccess    string
	StorageSecret    string
	StorageBucket    string
	AnthropicKey     string
	AnthropicFast    string
	AnthropicCapable string
	OllamaBaseURL    string
	OllamaFast       string
	OllamaCapable    string
	AdminEmail       string
}

func Load() Config {
	return Config{
		Environment:      env("APP_ENV", "development"),
		Address:          env("APP_ADDR", ":8080"),
		DatabaseURL:      env("DATABASE_URL", ""),
		CORSOrigins:      envList("CORS_ORIGINS", []string{"http://localhost:3000"}),
		JWTSecret:        env("JWT_SECRET", "change-me-at-least-32-chars-long!!"),
		FrontendURL:      env("FRONTEND_URL", "http://localhost:3000"),
		SidecarURL:       env("SIDECAR_URL", "http://localhost:3001"),
		StorageEndpoint:  env("STORAGE_ENDPOINT", ""),
		StorageAccess:    env("STORAGE_ACCESS_KEY", ""),
		StorageSecret:    env("STORAGE_SECRET_KEY", ""),
		StorageBucket:    env("STORAGE_BUCKET", "seed-assets"),
		AnthropicKey:     env("ANTHROPIC_API_KEY", ""),
		AnthropicFast:    env("ANTHROPIC_FAST_MODEL", "claude-haiku-4-5-20251001"),
		AnthropicCapable: env("ANTHROPIC_CAPABLE_MODEL", "claude-sonnet-4-6"),
		OllamaBaseURL:    env("OLLAMA_BASE_URL", "http://localhost:11434"),
		OllamaFast:       env("OLLAMA_FAST_MODEL", "qwen3:0.6b"),
		OllamaCapable:    env("OLLAMA_CAPABLE_MODEL", "qwen3:8b"),
		AdminEmail:       env("ADMIN_EMAIL", ""),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envList(key string, fallback []string) []string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return append([]string(nil), fallback...)
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return append([]string(nil), fallback...)
	}
	return out
}
