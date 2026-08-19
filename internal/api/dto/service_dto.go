package dto

import "time"

// Service Management Requests
type CreateServiceRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
}

type UpdateServiceStatusRequest struct {
	IsActive *bool `json:"is_active" validate:"required"`
}

type AssignUserRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

type CreateSecretRequest struct {
	Name      string     `json:"name" validate:"required,min=2,max=100"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type CreatePermissionRequest struct {
	TargetServiceID string `json:"target_service_id" validate:"required,uuid"`
}

// M2M Auth & Verification Requests
type ServiceTokenRequest struct {
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
}

type VerifyServiceTokenRequest struct {
	TargetClientID     string `json:"target_client_id" validate:"required"`
	TargetClientSecret string `json:"target_client_secret" validate:"required"`
	CallerToken        string `json:"caller_token" validate:"required"`
}

// Responses
type PaginatedServicesResponse struct {
	Services   any   `json:"services"`
	TotalCount int64 `json:"total_count"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
}
