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
	ID           string    `db:"id"            json:"id"`
	Email        string    `db:"email"         json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Role         Role      `db:"role"          json:"role"`
	Plan         Plan      `db:"plan"          json:"plan"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"    json:"updated_at"`
}
