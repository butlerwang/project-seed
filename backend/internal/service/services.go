package service

import (
	"github.com/butlerwang/project-seed/backend/internal/config"
	"github.com/butlerwang/project-seed/backend/internal/repository"
)

type Services struct {
	Auth *AuthService
}

func New(cfg config.Config, repos repository.Repository) *Services {
	return &Services{
		Auth: NewAuthService(cfg, repos),
	}
}
