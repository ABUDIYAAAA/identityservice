package router

import (
	"log/slog"
	"net/http"

	"devclub.com/identity/internal/api/config"
	"devclub.com/identity/internal/api/handlers"
	customMiddleware "devclub.com/identity/internal/api/middlewares"
	"devclub.com/identity/internal/api/services"
	"devclub.com/identity/internal/database"
	"devclub.com/identity/internal/mailer"
	"devclub.com/identity/pkg/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthMiddleware interface for route protection
type AuthMiddleware interface {
	Authenticate(next http.Handler) http.Handler
	RequireAdmin(next http.Handler) http.Handler
}

// NewRouter wires dependencies, initializes services/handlers, and registers all domain routes.
func NewRouter(
	cfg *config.Config,
	pool *pgxpool.Pool,
	mailClient *mailer.Mailer,
	jwtManager *utils.JWTManager,
	infoLogger *slog.Logger,
	warnLogger *slog.Logger,
	errorLogger *slog.Logger,
) *chi.Mux {
	r := chi.NewMux()

	// 1. Core Top-Level Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(customMiddleware.RequestLogger(infoLogger))
	r.Use(customMiddleware.Recoverer(errorLogger))

	// 2. Repositories
	authRepo := database.NewAuthRepository(pool)

	// 3. Services
	authSvc := services.NewAuthService(
		authRepo,
		jwtManager,
		mailClient,
		cfg,
		infoLogger,
		warnLogger,
		errorLogger,
	)

	// 4. Middlewares & Handlers
	authMW := customMiddleware.NewAuthMiddleware(jwtManager, pool, warnLogger, errorLogger)
	authH := handlers.NewAuthHandler(authSvc, jwtManager)
	userH := handlers.NewUserHandler(authSvc, authRepo)
	sessionH := handlers.NewSessionHandler(authSvc)

	// 5. Mount API Version Group
	r.Route("/api/v1", func(api chi.Router) {
		RegisterAuthRoutes(api, authH, authMW)
		RegisterUserRoutes(api, userH, authMW)
		RegisterSessionRoutes(api, sessionH, authMW)
	})

	return r
}
