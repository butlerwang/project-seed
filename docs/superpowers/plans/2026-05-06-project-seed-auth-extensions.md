# project-seed Auth Extensions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add email verification, password reset, and Google OAuth to project-seed so every cloned project has production-ready auth out of the box.

**Architecture:** Three independent auth flows built on top of the existing `AuthService`. A new `mailer` package (interface + SMTP impl + mock) handles all email. A single `email_tokens` DB table stores both verification and reset tokens (distinguished by `type`). Google OAuth is stateless — state param lives in a short-lived cookie, user info is fetched from Google's userinfo endpoint, then we find-or-create the user and issue our own JWT. No passwords for OAuth users (empty `password_hash`).

**Tech Stack:** Go stdlib `net/smtp`, `golang.org/x/oauth2` + `golang.org/x/oauth2/google`, goose migrations, existing chi router + repository pattern.

---

## File Structure

```
backend/
  internal/
    mailer/
      mailer.go          NEW — Mailer interface, SMTPMailer, MockMailer
    model/
      user.go            MODIFY — add EmailVerified bool field
      token.go           NEW — EmailToken model + TokenType consts
    repository/
      interface.go       MODIFY — add 5 token + UpdateUser methods
      postgres.go        MODIFY — implement new methods
      memory.go          MODIFY — implement new methods
    config/
      config.go          MODIFY — add AppURL, SMTP*, Google* fields
    service/
      auth.go            MODIFY — add VerifyEmail, ForgotPassword, ResetPassword, FindOrCreateGoogleUser
      auth_test.go       MODIFY — add tests for new methods
      services.go        MODIFY — inject Mailer
    handler/
      auth.go            MODIFY — add verifyEmail, forgotPassword, resetPassword, googleRedirect, googleCallback handlers
      router.go          MODIFY — register 5 new routes
  migrations/
    0002_auth_tokens.sql NEW — email_tokens table + email_verified column
.env.example             MODIFY — add SMTP_*, GOOGLE_*, APP_URL vars
```

---

### Task 1: Add dependencies and mailer package

**Files:**
- Modify: `backend/go.mod`
- Create: `backend/internal/mailer/mailer.go`

- [ ] **Step 1: Add golang.org/x/oauth2**

```bash
cd backend && go get golang.org/x/oauth2 golang.org/x/oauth2/google
```

Expected: go.mod updated, `golang.org/x/oauth2` and `golang.org/x/oauth2/google` appear in `require` block.

- [ ] **Step 2: Create mailer package**

Create `backend/internal/mailer/mailer.go`:

```go
package mailer

import (
	"fmt"
	"net/smtp"
	"strings"
)

// Mailer sends transactional emails.
type Mailer interface {
	Send(to, subject, htmlBody string) error
}

// SMTPMailer sends via SMTP. Works with SendGrid, Resend, Mailgun, and plain SMTP.
// Set SMTP_USER to the API key for SendGrid/Resend (username is "apikey" for SendGrid).
type SMTPMailer struct {
	Host string // e.g. "smtp.sendgrid.net"
	Port string // e.g. "587"
	From string // display + sender address, e.g. "MyApp <noreply@myapp.com>"
	User string // SMTP username
	Pass string // SMTP password / API key
}

func (m *SMTPMailer) Send(to, subject, htmlBody string) error {
	addr := m.Host + ":" + m.Port
	auth := smtp.PlainAuth("", m.User, m.Pass, m.Host)
	msg := strings.Join([]string{
		fmt.Sprintf("From: %s", m.From),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		htmlBody,
	}, "\r\n")
	return smtp.SendMail(addr, auth, m.From, []string{to}, []byte(msg))
}

// MockMailer records sent mail for use in tests. Thread-safe for concurrent use.
type MockMailer struct {
	Sent []SentMail
}

// SentMail is one recorded email.
type SentMail struct {
	To      string
	Subject string
	Body    string
}

func (m *MockMailer) Send(to, subject, htmlBody string) error {
	m.Sent = append(m.Sent, SentMail{To: to, Subject: subject, Body: htmlBody})
	return nil
}
```

- [ ] **Step 3: Verify it compiles**

```bash
cd backend && go build ./internal/mailer/...
```

Expected: no output (clean build).

- [ ] **Step 4: Commit**

```bash
cd backend
git add go.mod go.sum internal/mailer/
git commit -m "feat: add mailer package and golang.org/x/oauth2 dependency"
```

---

### Task 2: Add EmailToken model and update User model

**Files:**
- Create: `backend/internal/model/token.go`
- Modify: `backend/internal/model/user.go`

- [ ] **Step 1: Create token model**

Create `backend/internal/model/token.go`:

```go
package model

import "time"

// TokenType distinguishes email verification tokens from password reset tokens.
type TokenType string

const (
	TokenVerify TokenType = "verify"
	TokenReset  TokenType = "reset"
)

// EmailToken is stored hashed in the DB. The raw token is sent to the user via email.
type EmailToken struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	TokenHash string    `db:"token_hash"`
	Type      TokenType `db:"type"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}
```

- [ ] **Step 2: Add EmailVerified to User**

Replace `backend/internal/model/user.go` with:

```go
package model

import "time"

type Role string
type Plan string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"

	PlanFree Plan = "free"
	PlanPro  Plan = "pro"
)

type User struct {
	ID            string    `db:"id"             json:"id"`
	Email         string    `db:"email"          json:"email"`
	PasswordHash  string    `db:"password_hash"  json:"-"`
	Role          Role      `db:"role"           json:"role"`
	Plan          Plan      `db:"plan"           json:"plan"`
	EmailVerified bool      `db:"email_verified" json:"email_verified"`
	CreatedAt     time.Time `db:"created_at"     json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"     json:"updated_at"`
}
```

- [ ] **Step 3: Verify models compile**

```bash
cd backend && go build ./internal/model/...
```

Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add internal/model/
git commit -m "feat: add EmailToken model and EmailVerified field on User"
```

---

### Task 3: DB migration — email_tokens table

**Files:**
- Create: `backend/migrations/0002_auth_tokens.sql`

- [ ] **Step 1: Write migration**

Create `backend/migrations/0002_auth_tokens.sql`:

```sql
-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE email_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR NOT NULL UNIQUE,
    type        TEXT NOT NULL CHECK (type IN ('verify', 'reset')),
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX email_tokens_user_id_type_idx ON email_tokens(user_id, type);
CREATE INDEX email_tokens_expires_at_idx ON email_tokens(expires_at);

-- +goose Down
DROP TABLE IF EXISTS email_tokens;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified;
```

- [ ] **Step 2: Run against local DB to verify**

```bash
cd backend && make migrate
```

Expected: `OK    0002_auth_tokens.sql` in output.

- [ ] **Step 3: Commit**

```bash
git add migrations/
git commit -m "feat: add email_tokens table and email_verified column"
```

---

### Task 4: Extend repository interface and implementations

**Files:**
- Modify: `backend/internal/repository/interface.go`
- Modify: `backend/internal/repository/postgres.go`
- Modify: `backend/internal/repository/memory.go`

- [ ] **Step 1: Extend interface**

Replace `backend/internal/repository/interface.go` with:

```go
package repository

import "github.com/butlerwang/project-seed/backend/internal/model"

type Repository interface {
	// Users
	CreateUser(u model.User) error
	GetUserByID(id string) (model.User, error)
	GetUserByEmail(email string) (model.User, error)
	UpdateUser(u model.User) error
	ListUsers(limit, offset int) ([]model.User, error)

	// Sessions (refresh tokens)
	CreateSession(s model.Session) error
	GetSessionByTokenHash(hash string) (model.Session, error)
	DeleteSession(id string) error
	DeleteSessionsByUserID(userID string) error

	// EmailTokens (verify + reset)
	CreateEmailToken(t model.EmailToken) error
	GetEmailToken(tokenHash string, tokenType model.TokenType) (model.EmailToken, error)
	DeleteEmailToken(id string) error
	DeleteEmailTokensByUser(userID string, tokenType model.TokenType) error
}
```

- [ ] **Step 2: Implement new methods in PostgresRepository**

Add to `backend/internal/repository/postgres.go` (append before the closing of the file):

```go
func (r *PostgresRepository) UpdateUser(u model.User) error {
	_, err := r.db.NamedExec(`
		UPDATE users SET email=:email, password_hash=:password_hash, role=:role,
		plan=:plan, email_verified=:email_verified, updated_at=:updated_at
		WHERE id=:id`, u)
	return err
}

func (r *PostgresRepository) CreateEmailToken(t model.EmailToken) error {
	_, err := r.db.NamedExec(`
		INSERT INTO email_tokens (id, user_id, token_hash, type, expires_at, created_at)
		VALUES (:id, :user_id, :token_hash, :type, :expires_at, :created_at)`, t)
	return err
}

func (r *PostgresRepository) GetEmailToken(tokenHash string, tokenType model.TokenType) (model.EmailToken, error) {
	var t model.EmailToken
	err := r.db.Get(&t, `
		SELECT * FROM email_tokens
		WHERE token_hash=$1 AND type=$2 AND expires_at > now()`, tokenHash, string(tokenType))
	if err != nil {
		return model.EmailToken{}, mapErr(err)
	}
	return t, nil
}

func (r *PostgresRepository) DeleteEmailToken(id string) error {
	_, err := r.db.Exec(`DELETE FROM email_tokens WHERE id=$1`, id)
	return err
}

func (r *PostgresRepository) DeleteEmailTokensByUser(userID string, tokenType model.TokenType) error {
	_, err := r.db.Exec(`DELETE FROM email_tokens WHERE user_id=$1 AND type=$2`, userID, string(tokenType))
	return err
}
```

- [ ] **Step 3: Implement new methods in MemoryRepository**

Add to `backend/internal/repository/memory.go`:

First, add `tokens map[string]model.EmailToken` and `byTokenHash map[string]string` fields to the struct and `NewMemoryRepository()`:

```go
// Replace NewMemoryRepository with:
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:       make(map[string]model.User),
		byEmail:     make(map[string]string),
		sessions:    make(map[string]model.Session),
		byToken:     make(map[string]string),
		tokens:      make(map[string]model.EmailToken),
		byTokenHash: make(map[string]string),
	}
}

// Add to MemoryRepository struct:
// tokens      map[string]model.EmailToken
// byTokenHash map[string]string  // tokenHash → id
```

Then append these methods:

```go
func (m *MemoryRepository) UpdateUser(u model.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[u.ID]; !ok {
		return ErrNotFound
	}
	m.users[u.ID] = u
	m.byEmail[u.Email] = u.ID
	return nil
}

func (m *MemoryRepository) CreateEmailToken(t model.EmailToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[t.ID] = t
	m.byTokenHash[t.TokenHash] = t.ID
	return nil
}

func (m *MemoryRepository) GetEmailToken(tokenHash string, tokenType model.TokenType) (model.EmailToken, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.byTokenHash[tokenHash]
	if !ok {
		return model.EmailToken{}, ErrNotFound
	}
	t := m.tokens[id]
	if t.Type != tokenType || t.ExpiresAt.Before(time.Now()) {
		return model.EmailToken{}, ErrNotFound
	}
	return t, nil
}

func (m *MemoryRepository) DeleteEmailToken(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tokens[id]
	if !ok {
		return ErrNotFound
	}
	delete(m.byTokenHash, t.TokenHash)
	delete(m.tokens, id)
	return nil
}

func (m *MemoryRepository) DeleteEmailTokensByUser(userID string, tokenType model.TokenType) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, t := range m.tokens {
		if t.UserID == userID && t.Type == tokenType {
			delete(m.byTokenHash, t.TokenHash)
			delete(m.tokens, id)
		}
	}
	return nil
}
```

- [ ] **Step 4: Verify compilation**

```bash
cd backend && go build ./...
```

Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add internal/repository/
git commit -m "feat: extend repository with token and UpdateUser methods"
```

---

### Task 5: Update config

**Files:**
- Modify: `backend/internal/config/config.go`
- Modify: `backend/.env.example`

- [ ] **Step 1: Add new config fields**

Add to the `Config` struct in `config.go`:

```go
// After existing fields, add:
AppURL             string
SMTPHost           string
SMTPPort           string
SMTPFrom           string
SMTPUser           string
SMTPPass           string
GoogleClientID     string
GoogleClientSecret string
```

Add to the `Load()` return value:

```go
AppURL:             env("APP_URL", "http://localhost:8080"),
SMTPHost:           env("SMTP_HOST", ""),
SMTPPort:           env("SMTP_PORT", "587"),
SMTPFrom:           env("SMTP_FROM", "noreply@example.com"),
SMTPUser:           env("SMTP_USER", ""),
SMTPPass:           env("SMTP_PASS", ""),
GoogleClientID:     env("GOOGLE_CLIENT_ID", ""),
GoogleClientSecret: env("GOOGLE_CLIENT_SECRET", ""),
```

- [ ] **Step 2: Update .env.example**

Add to `.env.example`:

```
# Email (SMTP — use SendGrid, Resend, Mailgun, or plain SMTP)
APP_URL=http://localhost:8080
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_FROM=MyApp <noreply@myapp.com>
SMTP_USER=apikey
SMTP_PASS=SG.your-sendgrid-key

# Google OAuth (create at console.cloud.google.com)
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-secret
```

- [ ] **Step 3: Build**

```bash
cd backend && go build ./...
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add internal/config/config.go .env.example
git commit -m "feat: add SMTP and Google OAuth config fields"
```

---

### Task 6: Auth service — email verification + password reset

**Files:**
- Modify: `backend/internal/service/auth.go`
- Modify: `backend/internal/service/services.go`

- [ ] **Step 1: Write failing tests**

Add to `backend/internal/service/auth_test.go`:

```go
func TestVerifyEmail(t *testing.T) {
	repos := repository.NewMemoryRepository()
	mock := &mailer.MockMailer{}
	svc := NewAuthService(testCfg(), repos, mock)

	res, err := svc.Register("user@example.com", "password123")
	require.NoError(t, err)
	require.Len(t, mock.Sent, 1, "should send verification email on register")
	assert.Contains(t, mock.Sent[0].Subject, "Verify")

	// Extract token from email body (the link contains ?token=<raw>)
	body := mock.Sent[0].Body
	idx := strings.Index(body, "?token=")
	require.Greater(t, idx, -1)
	rawToken := body[idx+7 : idx+7+43] // base64url 32 bytes ≈ 43 chars

	err = svc.VerifyEmail(rawToken)
	require.NoError(t, err)

	u, _ := repos.GetUserByID(res.User.ID)
	assert.True(t, u.EmailVerified)
}

func TestForgotAndResetPassword(t *testing.T) {
	repos := repository.NewMemoryRepository()
	mock := &mailer.MockMailer{}
	svc := NewAuthService(testCfg(), repos, mock)

	_, err := svc.Register("reset@example.com", "oldpass")
	require.NoError(t, err)

	mock.Sent = nil // clear register email
	err = svc.ForgotPassword("reset@example.com")
	require.NoError(t, err)
	require.Len(t, mock.Sent, 1, "should send reset email")

	body := mock.Sent[0].Body
	idx := strings.Index(body, "?token=")
	require.Greater(t, idx, -1)
	rawToken := body[idx+7 : idx+7+43]

	err = svc.ResetPassword(rawToken, "newpass456")
	require.NoError(t, err)

	_, err = svc.Login("reset@example.com", "newpass456")
	assert.NoError(t, err)

	_, err = svc.Login("reset@example.com", "oldpass")
	assert.Error(t, err)
}

func testCfg() config.Config {
	c := config.Config{}
	c.JWTSecret = "test-secret-at-least-32-chars-ok"
	c.AppURL = "http://localhost:8080"
	return c
}
```

- [ ] **Step 2: Run tests — expect failure**

```bash
cd backend && go test ./internal/service/... -v -run "TestVerifyEmail|TestForgotAndResetPassword"
```

Expected: compile error (methods don't exist yet).

- [ ] **Step 3: Inject Mailer into AuthService**

Update `backend/internal/service/auth.go` — change the struct and constructor:

```go
// Change AuthService struct to:
type AuthService struct {
	cfg    config.Config
	repos  repository.Repository
	mailer mailer.Mailer
}

// Change NewAuthService to:
func NewAuthService(cfg config.Config, repos repository.Repository, m mailer.Mailer) *AuthService {
	return &AuthService{cfg: cfg, repos: repos, mailer: m}
}
```

Add imports at the top of `auth.go`:
```go
import (
	// existing imports +
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/butlerwang/project-seed/backend/internal/mailer"
	"github.com/butlerwang/project-seed/backend/internal/model"
)
```

- [ ] **Step 4: Send verification email on Register**

In `auth.go`, update `Register()` to send verification email after creating the user. Add this at the end of the `Register` method, before `return s.issueResult(u)`:

```go
// Send verification email (non-blocking — failure doesn't abort registration)
if s.mailer != nil && s.cfg.SMTPHost != "" {
    _ = s.sendVerificationEmail(u)
}
```

Add the helper method:

```go
func (s *AuthService) sendVerificationEmail(u model.User) error {
	rawToken, err := s.createEmailToken(u.ID, model.TokenVerify, 24*time.Hour)
	if err != nil {
		return err
	}
	link := fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", s.cfg.AppURL, rawToken)
	body := fmt.Sprintf(`<p>Welcome! Please verify your email:</p><p><a href="%s">Verify Email</a></p><p>Link expires in 24 hours.</p>`, link)
	return s.mailer.Send(u.Email, "Verify your email", body)
}

// createEmailToken generates a raw token, hashes it, stores it, returns the raw token.
func (s *AuthService) createEmailToken(userID string, tokenType model.TokenType, ttl time.Duration) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	rawToken := base64.RawURLEncoding.EncodeToString(raw)
	hash := tokenHash(rawToken) // reuse existing tokenHash() from auth.go

	_ = s.repos.DeleteEmailTokensByUser(userID, tokenType) // invalidate previous

	t := model.EmailToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: hash,
		Type:      tokenType,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now().UTC(),
	}
	return rawToken, s.repos.CreateEmailToken(t)
}
```

- [ ] **Step 5: Add VerifyEmail, ForgotPassword, ResetPassword methods**

Append to `auth.go`:

```go
func (s *AuthService) VerifyEmail(rawToken string) error {
	hash := tokenHash(rawToken)
	tok, err := s.repos.GetEmailToken(hash, model.TokenVerify)
	if err != nil {
		return errors.New("invalid or expired verification link")
	}
	u, err := s.repos.GetUserByID(tok.UserID)
	if err != nil {
		return err
	}
	u.EmailVerified = true
	u.UpdatedAt = time.Now().UTC()
	if err := s.repos.UpdateUser(u); err != nil {
		return err
	}
	return s.repos.DeleteEmailToken(tok.ID)
}

func (s *AuthService) ForgotPassword(email string) error {
	u, err := s.repos.GetUserByEmail(email)
	if err != nil {
		return nil // don't reveal whether email exists
	}
	rawToken, err := s.createEmailToken(u.ID, model.TokenReset, time.Hour)
	if err != nil {
		return err
	}
	link := fmt.Sprintf("%s/reset-password?token=%s", strings.TrimRight(s.cfg.AppURL, "/"), rawToken)

	// AppURL here is the frontend URL for the link but backend URL for the API.
	// Use FrontendURL if set.
	if s.cfg.FrontendURL != "" {
		link = fmt.Sprintf("%s/reset-password?token=%s", strings.TrimRight(s.cfg.FrontendURL, "/"), rawToken)
	}
	body := fmt.Sprintf(`<p>Reset your password:</p><p><a href="%s">Reset Password</a></p><p>Link expires in 1 hour.</p>`, link)
	return s.mailer.Send(u.Email, "Reset your password", body)
}

func (s *AuthService) ResetPassword(rawToken, newPassword string) error {
	if len(newPassword) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	hash := tokenHash(rawToken)
	tok, err := s.repos.GetEmailToken(hash, model.TokenReset)
	if err != nil {
		return errors.New("invalid or expired reset link")
	}
	u, err := s.repos.GetUserByID(tok.UserID)
	if err != nil {
		return err
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(newHash)
	u.UpdatedAt = time.Now().UTC()
	if err := s.repos.UpdateUser(u); err != nil {
		return err
	}
	_ = s.repos.DeleteEmailToken(tok.ID)
	_ = s.repos.DeleteSessionsByUserID(u.ID) // invalidate all existing sessions
	return nil
}
```

- [ ] **Step 6: Update services.go to inject Mailer**

Replace `backend/internal/service/services.go`:

```go
package service

import (
	"github.com/butlerwang/project-seed/backend/internal/config"
	"github.com/butlerwang/project-seed/backend/internal/mailer"
	"github.com/butlerwang/project-seed/backend/internal/repository"
)

type Services struct {
	Auth *AuthService
}

func New(cfg config.Config, repos repository.Repository) *Services {
	var m mailer.Mailer
	if cfg.SMTPHost != "" {
		m = &mailer.SMTPMailer{
			Host: cfg.SMTPHost,
			Port: cfg.SMTPPort,
			From: cfg.SMTPFrom,
			User: cfg.SMTPUser,
			Pass: cfg.SMTPPass,
		}
	}
	return &Services{
		Auth: NewAuthService(cfg, repos, m),
	}
}
```

- [ ] **Step 7: Run tests — expect pass**

```bash
cd backend && go test ./internal/service/... -v -run "TestVerifyEmail|TestForgotAndResetPassword"
```

Expected: PASS for both tests.

- [ ] **Step 8: Run all tests**

```bash
cd backend && go test ./...
```

Expected: all pass.

- [ ] **Step 9: Commit**

```bash
git add internal/service/ internal/mailer/
git commit -m "feat: add email verification and password reset to AuthService"
```

---

### Task 7: Auth service — Google OAuth (find-or-create)

**Files:**
- Modify: `backend/internal/service/auth.go`

- [ ] **Step 1: Write failing test**

Add to `auth_test.go`:

```go
func TestFindOrCreateGoogleUser(t *testing.T) {
	repos := repository.NewMemoryRepository()
	svc := NewAuthService(testCfg(), repos, &mailer.MockMailer{})

	// First call creates the user
	res, err := svc.FindOrCreateGoogleUser("googleuser@example.com")
	require.NoError(t, err)
	assert.Equal(t, "googleuser@example.com", res.User.Email)
	assert.True(t, res.User.EmailVerified)

	// Second call with same email returns existing user
	res2, err := svc.FindOrCreateGoogleUser("googleuser@example.com")
	require.NoError(t, err)
	assert.Equal(t, res.User.ID, res2.User.ID)
}
```

- [ ] **Step 2: Run — expect failure**

```bash
cd backend && go test ./internal/service/... -v -run TestFindOrCreateGoogleUser
```

Expected: compile error (method missing).

- [ ] **Step 3: Implement FindOrCreateGoogleUser**

Append to `backend/internal/service/auth.go`:

```go
// FindOrCreateGoogleUser looks up a user by email (from Google's userinfo),
// creates them if new (no password, email pre-verified), then issues a JWT.
func (s *AuthService) FindOrCreateGoogleUser(email string) (AuthResult, error) {
	u, err := s.repos.GetUserByEmail(email)
	if err != nil {
		// New user — create without password
		now := time.Now().UTC()
		u = model.User{
			ID:            uuid.NewString(),
			Email:         email,
			PasswordHash:  "", // no password for OAuth users
			Role:          model.RoleUser,
			Plan:          model.PlanFree,
			EmailVerified: true, // Google guarantees email ownership
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := s.repos.CreateUser(u); err != nil {
			return AuthResult{}, err
		}
	}
	return s.issueResult(u)
}
```

- [ ] **Step 4: Run test — expect pass**

```bash
cd backend && go test ./internal/service/... -v -run TestFindOrCreateGoogleUser
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/service/
git commit -m "feat: add FindOrCreateGoogleUser to AuthService"
```

---

### Task 8: HTTP handlers — verify email, forgot/reset password, Google OAuth

**Files:**
- Modify: `backend/internal/handler/auth.go`
- Modify: `backend/internal/handler/router.go`

- [ ] **Step 1: Add email verification and password reset handlers**

Append to `backend/internal/handler/auth.go`:

```go
func (h *authHandler) verifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeErr(w, http.StatusBadRequest, "missing token")
		return
	}
	if err := h.svc.Auth.VerifyEmail(token); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	// Redirect to frontend success page
	http.Redirect(w, r, h.frontendURL+"/login?verified=1", http.StatusFound)
}

func (h *authHandler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	// Always 204 — don't reveal whether email exists
	_ = h.svc.Auth.ForgotPassword(body.Email)
	w.WriteHeader(http.StatusNoContent)
}

func (h *authHandler) resetPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.svc.Auth.ResetPassword(body.Token, body.Password); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

Update `authHandler` struct to include `frontendURL`:

```go
type authHandler struct {
	svc         *service.Services
	frontendURL string
}
```

- [ ] **Step 2: Add Google OAuth handlers**

Append to `auth.go` (requires `golang.org/x/oauth2` and `golang.org/x/oauth2/google`):

```go
import (
	// add to existing imports:
	"encoding/json"
	"io"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func googleOAuthConfig(cfg config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.AppURL + "/api/v1/auth/google/callback",
		Scopes:       []string{"openid", "email"},
		Endpoint:     google.Endpoint,
	}
}

func (h *authHandler) googleRedirect(w http.ResponseWriter, r *http.Request) {
	if h.googleCfg == nil {
		writeErr(w, http.StatusNotImplemented, "google oauth not configured")
		return
	}
	state := generateState()
	http.SetCookie(w, &http.Cookie{
		Name: "oauth_state", Value: state, HttpOnly: true,
		Secure: true, SameSite: http.SameSiteLaxMode, Path: "/", MaxAge: 300,
	})
	http.Redirect(w, r, h.googleCfg.AuthCodeURL(state), http.StatusFound)
}

func (h *authHandler) googleCallback(w http.ResponseWriter, r *http.Request) {
	if h.googleCfg == nil {
		writeErr(w, http.StatusNotImplemented, "google oauth not configured")
		return
	}
	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value != r.URL.Query().Get("state") {
		writeErr(w, http.StatusBadRequest, "invalid oauth state")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", MaxAge: -1, Path: "/"})

	token, err := h.googleCfg.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "oauth exchange failed")
		return
	}
	email, err := fetchGoogleEmail(r.Context(), h.googleCfg, token)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "failed to fetch user info")
		return
	}
	result, err := h.svc.Auth.FindOrCreateGoogleUser(email)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "failed to create user")
		return
	}
	setRefreshCookie(w, result.RefreshToken)
	// Redirect to frontend dashboard with access token in query (SPA picks it up)
	http.Redirect(w, r, h.frontendURL+"/dashboard?token="+result.Token, http.StatusFound)
}

func fetchGoogleEmail(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (string, error) {
	client := cfg.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var info struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(data, &info); err != nil || info.Email == "" {
		return "", errors.New("no email in google response")
	}
	return info.Email, nil
}

func generateState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
```

Update `authHandler` struct (consolidate):

```go
type authHandler struct {
	svc         *service.Services
	frontendURL string
	googleCfg   *oauth2.Config
}
```

- [ ] **Step 3: Register new routes in router.go**

Update `NewRouter` in `backend/internal/handler/router.go`:

```go
// Replace the auth and admin handler construction with:
auth := &authHandler{
	svc:         svc,
	frontendURL: cfg.FrontendURL,
	googleCfg:   googleOAuthConfig(cfg), // returns nil if ClientID is empty — guarded in handler
}
admin := &adminHandler{repos: repos}

// In the r.Route("/api/v1", ...) block, add after existing auth routes:
r.Get("/auth/verify-email", auth.verifyEmail)
r.Post("/auth/forgot-password", auth.forgotPassword)
r.Post("/auth/reset-password", auth.resetPassword)
r.Get("/auth/google", auth.googleRedirect)
r.Get("/auth/google/callback", auth.googleCallback)
```

Update `googleOAuthConfig` to return nil when not configured:

```go
func googleOAuthConfig(cfg config.Config) *oauth2.Config {
	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
		return nil
	}
	return &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.AppURL + "/api/v1/auth/google/callback",
		Scopes:       []string{"openid", "email"},
		Endpoint:     google.Endpoint,
	}
}
```

- [ ] **Step 4: Build everything**

```bash
cd backend && go build ./...
```

Expected: no errors.

- [ ] **Step 5: Run all tests**

```bash
cd backend && go test ./...
```

Expected: all pass.

- [ ] **Step 6: Smoke test locally**

```bash
make up
curl -s http://localhost:8080/health | jq
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' | jq
curl -s -X POST http://localhost:8080/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}' -o /dev/null -w "%{http_code}"
```

Expected: health 200, register 201 with token, forgot-password 204.

- [ ] **Step 7: Commit**

```bash
git add internal/handler/ internal/service/ internal/config/
git commit -m "feat: add email verification, password reset, and Google OAuth endpoints"
```

---

### Task 9: Frontend — forgot password and reset password pages

**Files:**
- Create: `frontend/app/(auth)/forgot-password/page.tsx`
- Create: `frontend/app/(auth)/reset-password/page.tsx`
- Modify: `frontend/lib/api.ts`

- [ ] **Step 1: Add auth API methods to api.ts**

Add to the `api.auth` object in `frontend/lib/api.ts`:

```typescript
forgotPassword: (email: string) =>
  request<void>("/api/v1/auth/forgot-password", {
    method: "POST",
    body: JSON.stringify({ email }),
  }),
resetPassword: (token: string, password: string) =>
  request<void>("/api/v1/auth/reset-password", {
    method: "POST",
    body: JSON.stringify({ token, password }),
  }),
```

- [ ] **Step 2: Create forgot-password page**

Create `frontend/app/(auth)/forgot-password/page.tsx`:

```tsx
"use client";
import { useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [sent, setSent] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      await api.auth.forgotPassword(email);
      setSent(true);
    } catch {
      setError("Something went wrong. Try again.");
    }
  }

  if (sent) {
    return (
      <div className="text-center">
        <h1 className="text-2xl font-bold mb-2">Check your email</h1>
        <p className="text-gray-500">If that address exists, we sent a reset link.</p>
        <Link href="/login" className="mt-4 inline-block text-sm text-blue-600 hover:underline">
          Back to login
        </Link>
      </div>
    );
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Forgot password</h1>
      <form onSubmit={handleSubmit} className="space-y-4">
        <input
          type="email" required placeholder="Email" value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="w-full px-4 py-2 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-black"
        />
        {error && <p className="text-red-500 text-sm">{error}</p>}
        <button type="submit"
          className="w-full bg-black text-white py-2 rounded-xl font-medium hover:bg-gray-800">
          Send reset link
        </button>
      </form>
      <p className="mt-4 text-center text-sm text-gray-500">
        <Link href="/login" className="text-blue-600 hover:underline">Back to login</Link>
      </p>
    </div>
  );
}
```

- [ ] **Step 3: Create reset-password page**

Create `frontend/app/(auth)/reset-password/page.tsx`:

```tsx
"use client";
import { useState, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { api } from "@/lib/api";

function ResetForm() {
  const router = useRouter();
  const params = useSearchParams();
  const token = params.get("token") ?? "";
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    if (password.length < 8) {
      setError("Password must be at least 8 characters.");
      return;
    }
    try {
      await api.auth.resetPassword(token, password);
      router.push("/login?reset=1");
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Invalid or expired link.");
    }
  }

  if (!token) {
    return <p className="text-red-500">Missing reset token. Check your email link.</p>;
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Reset password</h1>
      <form onSubmit={handleSubmit} className="space-y-4">
        <input
          type="password" required placeholder="New password (8+ chars)" value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="w-full px-4 py-2 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-black"
        />
        {error && <p className="text-red-500 text-sm">{error}</p>}
        <button type="submit"
          className="w-full bg-black text-white py-2 rounded-xl font-medium hover:bg-gray-800">
          Set new password
        </button>
      </form>
    </div>
  );
}

export default function ResetPasswordPage() {
  return <Suspense><ResetForm /></Suspense>;
}
```

- [ ] **Step 4: Build frontend**

```bash
cd frontend && npm run build 2>&1 | tail -20
```

Expected: build succeeds, no TypeScript errors.

- [ ] **Step 5: Commit**

```bash
git add frontend/app/(auth)/forgot-password/ frontend/app/(auth)/reset-password/ frontend/lib/api.ts
git commit -m "feat: add forgot-password and reset-password pages"
```

---

### Task 10: Final integration test and PR

- [ ] **Step 1: Run full test suite**

```bash
cd backend && go test ./... -count=1
```

Expected: all pass.

- [ ] **Step 2: Build check**

```bash
cd backend && go build ./... && cd ../frontend && npm run build 2>&1 | tail -5
```

Expected: both succeed.

- [ ] **Step 3: Final commit**

```bash
git add .
git commit -m "feat: auth extensions complete — email verify, password reset, Google OAuth"
```
