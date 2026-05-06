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
