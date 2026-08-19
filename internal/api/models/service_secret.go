package models

import "time"

type ServiceSecret struct {
	ID           string     `json:"id"`
	ServiceID    string     `json:"service_id"`
	Name         string     `json:"name"`
	SecretPrefix string     `json:"secret_prefix"`
	SecretHash   string     `json:"-"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
