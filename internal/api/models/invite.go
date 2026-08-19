package models

import "time"

type UserInvitation struct {
	ID        string
	Email     string
	Role      string
	TokenHash string
	InvitedBy string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}
