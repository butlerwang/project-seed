package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Environment         string
	Address             string
	DatabaseURL         string
	CORSOrigins         []string
	JWTSecret           string
	AppURL              string
	FrontendURL         string
	SidecarURL          string
	StorageEndpoint     string
	StorageAccess       string
	StorageSecret       string
	StorageBucket       string
	SMTPHost            string
	SMTPPort            string
	SMTPFrom            string
	SMTPUser            string
	SMTPPass            string
	GoogleClientID      string
	GoogleClientSecret  string
	StripeWebhookSecret string
	RateLimitRPS        float64
	RateLimitBurst      int
	AnthropicKey        string
	AnthropicFast       string
	AnthropicCapable    string
	OllamaBaseURL       string
	OllamaFast          string
	OllamaCapable       string
	AdminEmail          string
}

func Load() Config {
	return Config{
		Environment:         env("APP_ENV", "development"),
		Address:             env("APP_ADDR", ":8080"),
		DatabaseURL:         env("DATABASE_URL", ""),
		CORSOrigins:         envList("CORS_ORIGINS", []string{"http://localhost:3000"}),
		JWTSecret:           env("JWT_SECRET", "change-me-at-least-32-chars-long!!"),
		AppURL:              env("APP_URL", "http://localhost:8080"),
		FrontendURL:         env("FRONTEND_URL", "http://localhost:3000"),
		SidecarURL:          env("SIDECAR_URL", "http://localhost:3001"),
		StorageEndpoint:     env("STORAGE_ENDPOINT", ""),
		StorageAccess:       env("STORAGE_ACCESS_KEY", ""),
		StorageSecret:       env("STORAGE_SECRET_KEY", ""),
		StorageBucket:       env("STORAGE_BUCKET", "seed-assets"),
		StripeWebhookSecret: env("STRIPE_WEBHOOK_SECRET", ""),
		RateLimitRPS:        envFloat("RATE_LIMIT_RPS", 10),
		RateLimitBurst:      envInt("RATE_LIMIT_BURST", 20),
		SMTPHost:            env("SMTP_HOST", ""),
		SMTPPort:            env("SMTP_PORT", "587"),
		SMTPFrom:            env("SMTP_FROM", "noreply@example.com"),
		SMTPUser:            env("SMTP_USER", ""),
		SMTPPass:            env("SMTP_PASS", ""),
		GoogleClientID:      env("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:  env("GOOGLE_CLIENT_SECRET", ""),
		AnthropicKey:        env("ANTHROPIC_API_KEY", ""),
		AnthropicFast:       env("ANTHROPIC_FAST_MODEL", "claude-haiku-4-5-20251001"),
		AnthropicCapable:    env("ANTHROPIC_CAPABLE_MODEL", "claude-sonnet-4-6"),
		OllamaBaseURL:       env("OLLAMA_BASE_URL", "http://localhost:11434"),
		OllamaFast:          env("OLLAMA_FAST_MODEL", "qwen3:0.6b"),
		OllamaCapable:       env("OLLAMA_CAPABLE_MODEL", "qwen3:8b"),
		AdminEmail:          env("ADMIN_EMAIL", ""),
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

func envFloat(key string, fallback float64) float64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
