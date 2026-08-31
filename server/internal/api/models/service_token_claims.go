package models

type ServiceTokenClaims struct {
	ServiceID    string `json:"service_id"`
	ClientID     string `json:"client_id"`
	TokenVersion int    `json:"token_version"`
}
