package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateServerStartupAllowsDevelopmentDefaults(t *testing.T) {
	cfg := Config{Environment: "development"}
	assert.NoError(t, cfg.ValidateServerStartup())
}

func TestValidateServerStartupRejectsProductionDefaults(t *testing.T) {
	cfg := Config{
		Environment: "production",
		JWTSecret:   defaultJWTSecret,
	}
	assert.Error(t, cfg.ValidateServerStartup())
}

func TestValidateServerStartupAcceptsProductionConfig(t *testing.T) {
	cfg := Config{
		Environment: "production",
		DatabaseURL: "postgres://user:pass@example.com/db",
		JWTSecret:   "a-production-secret-at-least-32-chars",
	}
	assert.NoError(t, cfg.ValidateServerStartup())
}
