package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type SessionHandler interface {
	ListSessions(w http.ResponseWriter, r *http.Request)
	RevokeSession(w http.ResponseWriter, r *http.Request)
	RevokeOtherSessions(w http.ResponseWriter, r *http.Request)
	RevokeAllSessions(w http.ResponseWriter, r *http.Request)
}

func RegisterSessionRoutes(r chi.Router, sessionH SessionHandler, authMW AuthMiddleware) {
	r.Route("/sessions", func(r chi.Router) {
		r.Use(authMW.Authenticate)

		r.Get("/", sessionH.ListSessions)
		r.Delete("/{sessionId}", sessionH.RevokeSession)
		r.Post("/revoke-others", sessionH.RevokeOtherSessions)
		r.Post("/revoke-all", sessionH.RevokeAllSessions)
	})
}
