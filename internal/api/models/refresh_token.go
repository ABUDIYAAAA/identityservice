package models

import "time"

type RefreshTokenSession struct {
	ID        string
	UserID    string
	TokenHash string
	UserAgent string
	IPAddress string
	IsRevoked bool
	ExpiresAt time.Time
	CreatedAt time.Time
}
