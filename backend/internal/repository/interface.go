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

	// EmailTokens
	CreateEmailToken(t model.EmailToken) error
	GetEmailToken(tokenHash string, tokenType model.TokenType) (model.EmailToken, error)
	DeleteEmailToken(id string) error
	DeleteEmailTokensByUser(userID string, tokenType model.TokenType) error
}
