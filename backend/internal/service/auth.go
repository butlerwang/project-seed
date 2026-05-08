package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/butlerwang/project-seed/backend/internal/config"
	"github.com/butlerwang/project-seed/backend/internal/mailer"
	"github.com/butlerwang/project-seed/backend/internal/model"
	"github.com/butlerwang/project-seed/backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrRefreshExpired = errors.New("refresh token expired")

type AuthClaims struct {
	Subject   string
	Email     string
	Role      model.Role
	ExpiresAt time.Time
}

type AuthResult struct {
	Token        string     `json:"token"`
	User         model.User `json:"user"`
	RefreshToken string     `json:"-"`
}

type AuthService struct {
	cfg    config.Config
	repos  repository.Repository
	mailer mailer.Mailer
}

func NewAuthService(cfg config.Config, repos repository.Repository, m mailer.Mailer) *AuthService {
	return &AuthService{cfg: cfg, repos: repos, mailer: m}
}

func (s *AuthService) Register(email, password string) (AuthResult, error) {
	if email == "" || password == "" {
		return AuthResult{}, errors.New("email and password required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, err
	}
	now := time.Now().UTC()
	u := model.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         model.RoleUser,
		Plan:         model.PlanFree,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repos.CreateUser(u); err != nil {
		return AuthResult{}, err
	}
	if s.mailer != nil {
		_ = s.sendVerificationEmail(u)
	}
	return s.issueResult(u)
}

func (s *AuthService) Login(email, password string) (AuthResult, error) {
	u, err := s.repos.GetUserByEmail(email)
	if err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}
	return s.issueResult(u)
}

func (s *AuthService) Refresh(rawToken string) (AuthResult, error) {
	hash := tokenHash(rawToken)
	sess, err := s.repos.GetSessionByTokenHash(hash)
	if err != nil {
		return AuthResult{}, ErrRefreshExpired
	}
	u, err := s.repos.GetUserByID(sess.UserID)
	if err != nil {
		return AuthResult{}, ErrRefreshExpired
	}
	_ = s.repos.DeleteSession(sess.ID)
	return s.issueResult(u)
}

func (s *AuthService) Logout(rawToken string) error {
	hash := tokenHash(rawToken)
	sess, err := s.repos.GetSessionByTokenHash(hash)
	if err != nil {
		return nil
	}
	return s.repos.DeleteSession(sess.ID)
}

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
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return err
	}

	if s.mailer == nil {
		return nil
	}

	rawToken, err := s.createEmailToken(u.ID, model.TokenReset, time.Hour)
	if err != nil {
		return err
	}

	baseURL := strings.TrimRight(s.cfg.AppURL, "/")
	if s.cfg.FrontendURL != "" {
		baseURL = strings.TrimRight(s.cfg.FrontendURL, "/")
	}
	link := fmt.Sprintf("%s/reset-password?token=%s", baseURL, rawToken)
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
	_ = s.repos.DeleteSessionsByUserID(u.ID)
	return nil
}

// FindOrCreateGoogleUser creates a local account for a verified Google identity if needed.
func (s *AuthService) FindOrCreateGoogleUser(email string) (AuthResult, error) {
	u, err := s.repos.GetUserByEmail(email)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			return AuthResult{}, err
		}

		now := time.Now().UTC()
		u = model.User{
			ID:            uuid.NewString(),
			Email:         email,
			PasswordHash:  "",
			Role:          model.RoleUser,
			Plan:          model.PlanFree,
			EmailVerified: true,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := s.repos.CreateUser(u); err != nil {
			return AuthResult{}, err
		}
	}

	return s.issueResult(u)
}

func (s *AuthService) ParseToken(tokenString string) (AuthClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return []byte(s.cfg.JWTSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return AuthClaims{}, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return AuthClaims{}, errors.New("invalid claims")
	}
	exp, _ := claims.GetExpirationTime()
	sub, err := claims.GetSubject()
	if err != nil || sub == "" || exp == nil {
		return AuthClaims{}, errors.New("invalid claims")
	}
	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return AuthClaims{}, errors.New("invalid claims")
	}
	role, ok := claims["role"].(string)
	if !ok || role == "" {
		return AuthClaims{}, errors.New("invalid claims")
	}
	return AuthClaims{
		Subject:   sub,
		Email:     email,
		Role:      model.Role(role),
		ExpiresAt: exp.Time,
	}, nil
}

func (s *AuthService) issueResult(u model.User) (AuthResult, error) {
	accessToken, err := s.mintJWT(u)
	if err != nil {
		return AuthResult{}, err
	}
	rawRefresh, err := s.storeRefreshToken(u.ID)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{Token: accessToken, User: u, RefreshToken: rawRefresh}, nil
}

func (s *AuthService) mintJWT(u model.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   u.ID,
		"email": u.Email,
		"role":  string(u.Role),
		"iat":   now.Unix(),
		"exp":   now.Add(15 * time.Minute).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) storeRefreshToken(userID string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(raw)
	sess := model.Session{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: tokenHash(token),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	return token, s.repos.CreateSession(sess)
}

func (s *AuthService) sendVerificationEmail(u model.User) error {
	rawToken, err := s.createEmailToken(u.ID, model.TokenVerify, 24*time.Hour)
	if err != nil {
		return err
	}

	link := fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", strings.TrimRight(s.cfg.AppURL, "/"), rawToken)
	body := fmt.Sprintf(`<p>Welcome! Please verify your email:</p><p><a href="%s">Verify Email</a></p><p>Link expires in 24 hours.</p>`, link)
	return s.mailer.Send(u.Email, "Verify your email", body)
}

func (s *AuthService) createEmailToken(userID string, tokenType model.TokenType, ttl time.Duration) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	rawToken := base64.RawURLEncoding.EncodeToString(raw)
	hash := tokenHash(rawToken)

	if err := s.repos.DeleteEmailTokensByUser(userID, tokenType); err != nil {
		return "", err
	}

	t := model.EmailToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: hash,
		Type:      tokenType,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repos.CreateEmailToken(t); err != nil {
		return "", err
	}

	return rawToken, nil
}

func tokenHash(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
