package models

import "time"

type Service struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	ClientID     string    `json:"client_id"`
	IsActive     bool      `json:"is_active"`
	TokenVersion int       `json:"token_version"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
