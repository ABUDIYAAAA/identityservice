package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type DashboardHandler interface {
	GetDashboardStats(w http.ResponseWriter, r *http.Request)
}

func RegisterDashboardRoutes(r chi.Router, dashboardH DashboardHandler, authMW AuthMiddleware) {
	r.Route("/dashboard", func(r chi.Router) {
		r.Use(authMW.Authenticate)

		r.Get("/stats", dashboardH.GetDashboardStats)
	})
}
