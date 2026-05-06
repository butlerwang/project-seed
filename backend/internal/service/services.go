package service

import (
	"github.com/butlerwang/project-seed/backend/internal/config"
	"github.com/butlerwang/project-seed/backend/internal/llm"
	"github.com/butlerwang/project-seed/backend/internal/mailer"
	"github.com/butlerwang/project-seed/backend/internal/repository"
)

type Services struct {
	Auth *AuthService
	LLM  *llm.Router
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
		LLM:  llm.NewRouter(cfg),
	}
}
