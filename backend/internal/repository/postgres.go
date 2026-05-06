package repository

import (
	"github.com/butlerwang/project-seed/backend/internal/model"
	"github.com/jmoiron/sqlx"
)

type PostgresRepository struct{ db *sqlx.DB }

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateUser(u model.User) error {
	_, err := r.db.NamedExec(`
		INSERT INTO users (id, email, password_hash, role, plan, created_at, updated_at)
		VALUES (:id, :email, :password_hash, :role, :plan, :created_at, :updated_at)`, u)
	return err
}

func (r *PostgresRepository) GetUserByID(id string) (model.User, error) {
	var u model.User
	err := r.db.Get(&u, `SELECT * FROM users WHERE id = $1`, id)
	if err != nil {
		return model.User{}, mapErr(err)
	}
	return u, nil
}

func (r *PostgresRepository) GetUserByEmail(email string) (model.User, error) {
	var u model.User
	err := r.db.Get(&u, `SELECT * FROM users WHERE email = $1`, email)
	if err != nil {
		return model.User{}, mapErr(err)
	}
	return u, nil
}

func (r *PostgresRepository) UpdateUser(u model.User) error {
	_, err := r.db.NamedExec(`
		UPDATE users
		SET email=:email, password_hash=:password_hash, role=:role, plan=:plan,
		    email_verified=:email_verified, updated_at=:updated_at
		WHERE id=:id`, u)
	return err
}

func (r *PostgresRepository) ListUsers(limit, offset int) ([]model.User, error) {
	var users []model.User
	err := r.db.Select(&users, `SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	return users, err
}

func (r *PostgresRepository) CreateSession(s model.Session) error {
	_, err := r.db.NamedExec(`
		INSERT INTO sessions (id, user_id, token_hash, expires_at, created_at)
		VALUES (:id, :user_id, :token_hash, :expires_at, :created_at)`, s)
	return err
}

func (r *PostgresRepository) GetSessionByTokenHash(hash string) (model.Session, error) {
	var s model.Session
	err := r.db.Get(&s, `SELECT * FROM sessions WHERE token_hash = $1 AND expires_at > now()`, hash)
	if err != nil {
		return model.Session{}, mapErr(err)
	}
	return s, nil
}

func (r *PostgresRepository) DeleteSession(id string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE id = $1`, id)
	return err
}

func (r *PostgresRepository) DeleteSessionsByUserID(userID string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE user_id = $1`, userID)
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
		WHERE token_hash = $1 AND type = $2 AND expires_at > now()`, tokenHash, string(tokenType))
	if err != nil {
		return model.EmailToken{}, mapErr(err)
	}
	return t, nil
}

func (r *PostgresRepository) DeleteEmailToken(id string) error {
	_, err := r.db.Exec(`DELETE FROM email_tokens WHERE id = $1`, id)
	return err
}

func (r *PostgresRepository) DeleteEmailTokensByUser(userID string, tokenType model.TokenType) error {
	_, err := r.db.Exec(`DELETE FROM email_tokens WHERE user_id = $1 AND type = $2`, userID, string(tokenType))
	return err
}

func mapErr(err error) error {
	if err != nil && err.Error() == "sql: no rows in result set" {
		return ErrNotFound
	}
	return err
}
