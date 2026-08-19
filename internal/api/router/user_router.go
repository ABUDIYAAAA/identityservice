package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type UserHandler interface {
	GetProfile(w http.ResponseWriter, r *http.Request)
	ChangePassword(w http.ResponseWriter, r *http.Request)
	ListUsers(w http.ResponseWriter, r *http.Request)
	CreateInvite(w http.ResponseWriter, r *http.Request)
	UpdateRole(w http.ResponseWriter, r *http.Request)
	UpdateStatus(w http.ResponseWriter, r *http.Request)
	RemoveUser(w http.ResponseWriter, r *http.Request)
}

func RegisterUserRoutes(r chi.Router, userH UserHandler, authMW AuthMiddleware) {
	r.Route("/users", func(r chi.Router) {
		r.Use(authMW.Authenticate)

		// Current User Profile Actions
		r.Get("/me", userH.GetProfile)
		r.Put("/me/password", userH.ChangePassword)

		// Admin-Only Sub-Routes
		r.Group(func(r chi.Router) {
			r.Use(authMW.RequireAdmin)

			r.Get("/", userH.ListUsers)
			r.Post("/invite", userH.CreateInvite)
			r.Patch("/{id}/role", userH.UpdateRole)
			r.Patch("/{id}/status", userH.UpdateStatus)
			r.Delete("/{id}", userH.RemoveUser)
		})
	})
}
