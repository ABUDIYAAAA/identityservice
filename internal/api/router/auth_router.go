package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type AuthHandler interface {
	Login(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
	ForgotPassword(w http.ResponseWriter, r *http.Request)
	ResetPassword(w http.ResponseWriter, r *http.Request)
	AcceptInvite(w http.ResponseWriter, r *http.Request)
	GetInviteDetails(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}

func RegisterAuthRoutes(r chi.Router, authH AuthHandler, authMW AuthMiddleware) {
	r.Route("/auth", func(r chi.Router) {
		// Public Auth Endpoints
		r.Post("/login", authH.Login)
		r.Post("/refresh", authH.RefreshToken)
		r.Post("/forgot-password", authH.ForgotPassword)
		r.Post("/reset-password", authH.ResetPassword)

		// Invite Processing
		r.Get("/invite/{token}", authH.GetInviteDetails)
		r.Post("/invite/accept", authH.AcceptInvite)

		// Authenticated Auth Actions
		r.Group(func(r chi.Router) {
			r.Use(authMW.Authenticate)
			r.Post("/logout", authH.Logout)
		})
	})
}
