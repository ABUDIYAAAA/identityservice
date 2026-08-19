package models

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
	IsActive     bool
	TokenVersion int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
