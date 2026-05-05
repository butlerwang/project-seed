package repository

import "github.com/butlerwang/project-seed/backend/internal/model"

type Repository interface {
	// Users
	CreateUser(u model.User) error
	GetUserByID(id string) (model.User, error)
	GetUserByEmail(email string) (model.User, error)
	ListUsers(limit, offset int) ([]model.User, error)

	// Sessions (refresh tokens)
	CreateSession(s model.Session) error
	GetSessionByTokenHash(hash string) (model.Session, error)
	DeleteSession(id string) error
	DeleteSessionsByUserID(userID string) error
}
