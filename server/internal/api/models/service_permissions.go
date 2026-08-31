package models

import "time"

type ServicePermission struct {
	ID              string    `json:"id"`
	SourceServiceID string    `json:"source_service_id"`
	TargetServiceID string    `json:"target_service_id"`
	CreatedAt       time.Time `json:"created_at"`
}
