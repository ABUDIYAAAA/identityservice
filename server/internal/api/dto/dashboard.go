package dto

import "devclub.com/identity/internal/api/models"

type DashboardStatsResponse struct {
	ActiveUsersCount int64                  `json:"active_users_count"`
	ServicesCount    int64                  `json:"services_count"`
	FailedLogins24h  int64                  `json:"failed_logins_24h"`
	RecentAuditLogs  []models.AuditLogEvent `json:"recent_audit_logs"`
}
