package router

import (
	"net/http"
	"time"

	customMiddleware "devclub.com/identity/internal/api/middlewares"
	"devclub.com/identity/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type ServiceHandler interface {
	// Service CRUD
	CreateService(w http.ResponseWriter, r *http.Request)
	GetService(w http.ResponseWriter, r *http.Request)
	ListServices(w http.ResponseWriter, r *http.Request)
	UpdateServiceStatus(w http.ResponseWriter, r *http.Request)
	DeleteService(w http.ResponseWriter, r *http.Request)

	// User Assignments
	AssignUser(w http.ResponseWriter, r *http.Request)
	RemoveUser(w http.ResponseWriter, r *http.Request)

	// Multi-Secret Management
	GenerateSecret(w http.ResponseWriter, r *http.Request)
	ListSecrets(w http.ResponseWriter, r *http.Request)
	DeleteSecret(w http.ResponseWriter, r *http.Request)

	// Permissions (Inter-Service Links)
	AddPermission(w http.ResponseWriter, r *http.Request)
	RemovePermission(w http.ResponseWriter, r *http.Request)
	ListAllowedTargets(w http.ResponseWriter, r *http.Request)

	// M2M Auth & Verification
	GenerateToken(w http.ResponseWriter, r *http.Request)
	VerifyToken(w http.ResponseWriter, r *http.Request)
}

func RegisterServiceRoutes(r chi.Router, serviceH ServiceHandler, authMW AuthMiddleware) {
	// Rate limiters for M2M communication
	tokenLimiter := utils.NewRateLimiter(50.0, 100, 5*time.Minute)
	verifyLimiter := utils.NewRateLimiter(200.0, 500, 5*time.Minute)

	r.Route("/services", func(r chi.Router) {
		// ----------------------------------------------------
		// 1. M2M Machine-to-Machine Endpoints (Public + Rate-Limited)
		// ----------------------------------------------------
		r.With(customMiddleware.PerRouteRateLimit(tokenLimiter)).Post("/token", serviceH.GenerateToken)
		r.With(customMiddleware.PerRouteRateLimit(verifyLimiter)).Post("/verify", serviceH.VerifyToken)

		// ----------------------------------------------------
		// 2. Protected Service Management
		// ----------------------------------------------------
		r.Group(func(r chi.Router) {
			r.Use(authMW.Authenticate)

			// Service Listing & Details (Admins & Assigned Users)
			r.Get("/", serviceH.ListServices)
			r.Get("/{id}", serviceH.GetService)
			r.Get("/{id}/permissions", serviceH.ListAllowedTargets)

			// Secret Rotation & Management
			r.Route("/{id}/secrets", func(r chi.Router) {
				r.Get("/", serviceH.ListSecrets)
				r.Post("/", serviceH.GenerateSecret)
				r.Delete("/{secretId}", serviceH.DeleteSecret)
			})

			// ------------------------------------------------
			// 3. Admin-Only Service Operations
			// ------------------------------------------------
			r.Group(func(r chi.Router) {
				r.Use(authMW.RequireAdmin)

				r.Post("/", serviceH.CreateService)
				r.Patch("/{id}/status", serviceH.UpdateServiceStatus)
				r.Delete("/{id}", serviceH.DeleteService)

				// User Assignments
				r.Post("/{id}/assign", serviceH.AssignUser)
				r.Delete("/{id}/assign/{userId}", serviceH.RemoveUser)

				// Inter-Service Link Permissions
				r.Post("/{id}/permissions", serviceH.AddPermission)
				r.Delete("/{id}/permissions/{targetId}", serviceH.RemovePermission)
			})
		})
	})
}
