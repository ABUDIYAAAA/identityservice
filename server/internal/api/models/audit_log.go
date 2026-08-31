package models

import "time"

type AuditLogEvent struct {
	ID           string    `json:"id,omitempty"`
	ActionType   string    `json:"action_type"`
	ActorType    string    `json:"actor_type"` // "user", "service", "system"
	ActorID      string    `json:"actor_id,omitempty"`
	ServiceID    string    `json:"service_id,omitempty"` // Service-scoped event support
	BeforeState  any       `json:"before_state,omitempty"`
	AfterState   any       `json:"after_state,omitempty"`
	Status       string    `json:"status"` // "success", "failure"
	ErrorMessage string    `json:"error_message,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
