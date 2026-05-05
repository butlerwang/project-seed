package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/butlerwang/project-seed/backend/internal/config"
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
	cfg   config.Config
	repos repository.Repository
}

func NewAuthService(cfg config.Config, repos repository.Repository) *AuthService {
	return &AuthService{cfg: cfg, repos: repos}
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
	return AuthClaims{
		Subject:   claims["sub"].(string),
		Email:     claims["email"].(string),
		Role:      model.Role(claims["role"].(string)),
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

func tokenHash(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
