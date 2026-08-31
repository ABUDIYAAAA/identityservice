package handlers

import (
	"net/http"
	"strconv"

	"devclub.com/identity/internal/api/dto"
	"devclub.com/identity/internal/api/services"
	"devclub.com/identity/pkg/utils"
)

type AuditHandler struct {
	auditClient services.AuditClient
}

func NewAuditHandler(auditClient services.AuditClient) *AuditHandler {
	return &AuditHandler{auditClient: auditClient}
}

func (h *AuditHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	actionType := r.URL.Query().Get("action_type")
	actorID := r.URL.Query().Get("actor_id")
	serviceID := r.URL.Query().Get("service_id")

	logs, total, err := h.auditClient.ListAuditLogs(r.Context(), actionType, actorID, serviceID, limit, offset)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, "Audit logs retrieved successfully", dto.PaginatedAuditLogsResponse{
		AuditLogs:  logs,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	})
}
