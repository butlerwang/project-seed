package service_test

import (
	"testing"

	"github.com/butlerwang/project-seed/backend/internal/config"
	"github.com/butlerwang/project-seed/backend/internal/repository"
	"github.com/butlerwang/project-seed/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAuth() *service.AuthService {
	cfg := config.Config{JWTSecret: "test-secret-at-least-32-chars-long"}
	return service.NewAuthService(cfg, repository.NewMemoryRepository())
}

func TestRegisterAndLogin(t *testing.T) {
	auth := newTestAuth()

	result, err := auth.Register("test@example.com", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, result.Token)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, "test@example.com", result.User.Email)

	login, err := auth.Login("test@example.com", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, login.Token)
}

func TestLoginWrongPassword(t *testing.T) {
	auth := newTestAuth()
	_, _ = auth.Register("test@example.com", "correct")
	_, err := auth.Login("test@example.com", "wrong")
	assert.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestParseToken(t *testing.T) {
	auth := newTestAuth()
	result, err := auth.Register("parse@example.com", "password123")
	require.NoError(t, err)

	claims, err := auth.ParseToken(result.Token)
	require.NoError(t, err)
	assert.Equal(t, "parse@example.com", claims.Email)
}

func TestRefreshRotates(t *testing.T) {
	auth := newTestAuth()
	r1, err := auth.Register("refresh@example.com", "password123")
	require.NoError(t, err)

	r2, err := auth.Refresh(r1.RefreshToken)
	require.NoError(t, err)
	assert.NotEqual(t, r1.RefreshToken, r2.RefreshToken)

	// old token must be invalid after rotation
	_, err = auth.Refresh(r1.RefreshToken)
	assert.Error(t, err)
}

func TestDuplicateEmail(t *testing.T) {
	auth := newTestAuth()
	_, err := auth.Register("dupe@example.com", "password123")
	require.NoError(t, err)
	_, err = auth.Register("dupe@example.com", "other")
	assert.Error(t, err)
}
