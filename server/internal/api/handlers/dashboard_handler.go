package handlers

import (
	"net/http"

	"devclub.com/identity/internal/api/dto"
	"devclub.com/identity/internal/api/services"
	"devclub.com/identity/pkg/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DashboardHandler struct {
	db          *pgxpool.Pool
	auditClient services.AuditClient
}

func NewDashboardHandler(db *pgxpool.Pool, auditClient services.AuditClient) *DashboardHandler {
	return &DashboardHandler{
		db:          db,
		auditClient: auditClient,
	}
}

func (h *DashboardHandler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var activeUsersCount int64
	if h.db != nil {
		_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE is_active = true`).Scan(&activeUsersCount)
	}

	var servicesCount int64
	if h.db != nil {
		_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM services WHERE is_active = true`).Scan(&servicesCount)
	}

	var failedLogins24h int64
	if h.db != nil {
		_ = h.db.QueryRow(ctx, `
			SELECT COUNT(*) FROM audit_logs 
			WHERE action_type IN ('USER_LOGIN_FAILED', 'USER_LOGIN_LOCKED') 
			  AND created_at >= NOW() - INTERVAL '24 hours'
		`).Scan(&failedLogins24h)
	}

	recentLogs, _, _ := h.auditClient.ListAuditLogs(ctx, "", "", "", 5, 0)

	utils.Success(w, "Dashboard statistics retrieved", dto.DashboardStatsResponse{
		ActiveUsersCount: activeUsersCount,
		ServicesCount:    servicesCount,
		FailedLogins24h:  failedLogins24h,
		RecentAuditLogs:  recentLogs,
	})
}
