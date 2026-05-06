package service_test

import (
	"regexp"
	"testing"

	"github.com/butlerwang/project-seed/backend/internal/config"
	"github.com/butlerwang/project-seed/backend/internal/mailer"
	"github.com/butlerwang/project-seed/backend/internal/repository"
	"github.com/butlerwang/project-seed/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var tokenPattern = regexp.MustCompile(`token=([A-Za-z0-9_-]+)`)

func newTestAuth(repos repository.Repository, m *mailer.MockMailer) *service.AuthService {
	cfg := config.Config{
		JWTSecret:   "test-secret-at-least-32-chars-long",
		AppURL:      "http://localhost:8080",
		FrontendURL: "http://localhost:3000",
	}
	return service.NewAuthService(cfg, repos, m)
}

func TestRegisterAndLogin(t *testing.T) {
	auth := newTestAuth(repository.NewMemoryRepository(), &mailer.MockMailer{})

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
	auth := newTestAuth(repository.NewMemoryRepository(), &mailer.MockMailer{})
	_, _ = auth.Register("test@example.com", "correct")
	_, err := auth.Login("test@example.com", "wrong")
	assert.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestParseToken(t *testing.T) {
	auth := newTestAuth(repository.NewMemoryRepository(), &mailer.MockMailer{})
	result, err := auth.Register("parse@example.com", "password123")
	require.NoError(t, err)

	claims, err := auth.ParseToken(result.Token)
	require.NoError(t, err)
	assert.Equal(t, "parse@example.com", claims.Email)
}

func TestRefreshRotates(t *testing.T) {
	auth := newTestAuth(repository.NewMemoryRepository(), &mailer.MockMailer{})
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
	auth := newTestAuth(repository.NewMemoryRepository(), &mailer.MockMailer{})
	_, err := auth.Register("dupe@example.com", "password123")
	require.NoError(t, err)
	_, err = auth.Register("dupe@example.com", "other")
	assert.Error(t, err)
}

func TestVerifyEmail(t *testing.T) {
	repos := repository.NewMemoryRepository()
	mock := &mailer.MockMailer{}
	auth := newTestAuth(repos, mock)

	result, err := auth.Register("user@example.com", "password123")
	require.NoError(t, err)
	require.Len(t, mock.Sent, 1)
	assert.Contains(t, mock.Sent[0].Subject, "Verify")

	rawToken := extractToken(t, mock.Sent[0].Body)
	require.NoError(t, auth.VerifyEmail(rawToken))

	user, err := repos.GetUserByID(result.User.ID)
	require.NoError(t, err)
	assert.True(t, user.EmailVerified)
}

func TestForgotAndResetPassword(t *testing.T) {
	repos := repository.NewMemoryRepository()
	mock := &mailer.MockMailer{}
	auth := newTestAuth(repos, mock)

	_, err := auth.Register("reset@example.com", "oldpass123")
	require.NoError(t, err)

	mock.Sent = nil
	require.NoError(t, auth.ForgotPassword("reset@example.com"))
	require.Len(t, mock.Sent, 1)

	rawToken := extractToken(t, mock.Sent[0].Body)
	require.NoError(t, auth.ResetPassword(rawToken, "newpass456"))

	_, err = auth.Login("reset@example.com", "newpass456")
	assert.NoError(t, err)

	_, err = auth.Login("reset@example.com", "oldpass123")
	assert.Error(t, err)
}

func TestFindOrCreateGoogleUser(t *testing.T) {
	repos := repository.NewMemoryRepository()
	auth := newTestAuth(repos, &mailer.MockMailer{})

	first, err := auth.FindOrCreateGoogleUser("googleuser@example.com")
	require.NoError(t, err)
	assert.Equal(t, "googleuser@example.com", first.User.Email)
	assert.True(t, first.User.EmailVerified)

	second, err := auth.FindOrCreateGoogleUser("googleuser@example.com")
	require.NoError(t, err)
	assert.Equal(t, first.User.ID, second.User.ID)
}

func extractToken(t *testing.T, body string) string {
	t.Helper()
	match := tokenPattern.FindStringSubmatch(body)
	require.Len(t, match, 2, "expected token in email body")
	return match[1]
}
