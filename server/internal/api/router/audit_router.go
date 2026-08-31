package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type AuditHandler interface {
	ListAuditLogs(w http.ResponseWriter, r *http.Request)
}

func RegisterAuditRoutes(r chi.Router, auditH AuditHandler, authMW AuthMiddleware) {
	r.Route("/audit-logs", func(r chi.Router) {
		r.Use(authMW.Authenticate)
		r.Use(authMW.RequireAdmin) // Admin-only protection

		r.Get("/", auditH.ListAuditLogs)
	})
}
