package service

import "github.com/butlerwang/project-seed/backend/internal/config"

type Services struct {
	Auth *AuthService
}

func New(cfg config.Config, repos any) *Services {
	return &Services{}
}
