package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type HealthHandler interface {
	CheckHealth(w http.ResponseWriter, r *http.Request)
}

func RegisterHealthRoutes(r chi.Router, healthH HealthHandler) {
	r.Get("/health", healthH.CheckHealth)
}
