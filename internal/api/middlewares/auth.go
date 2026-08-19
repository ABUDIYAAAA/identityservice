package middlewares

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"devclub.com/identity/pkg/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Context keys
type contextKey string

const (
	UserContextKey   contextKey = "auth_user_claims"
	UserIDContextKey contextKey = "auth_user_id"
	UserRoleKey      contextKey = "auth_user_role"
)

type AuthMiddleware struct {
	jwtManager *utils.JWTManager
	db         *pgxpool.Pool
	warnLog    *slog.Logger
	errorLog   *slog.Logger
}

func NewAuthMiddleware(jwtManager *utils.JWTManager, db *pgxpool.Pool, warnLog, errorLog *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		db:         db,
		warnLog:    warnLog,
		errorLog:   errorLog,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.Unauthorized(w, "Authorization header is required")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			m.warnLog.Warn("malformed authorization header", "ip", r.RemoteAddr)
			utils.Unauthorized(w, "Invalid authorization format. Expected 'Bearer <token>'")
			return
		}

		tokenStr := parts[1]

		claims, err := m.jwtManager.VerifyAccessToken(tokenStr)
		if err != nil {
			if errors.Is(err, utils.ErrExpiredToken) {
				utils.Unauthorized(w, "Access token has expired")
				return
			}
			m.warnLog.Warn("invalid token signature or claims", "error", err, "ip", r.RemoteAddr)
			utils.Unauthorized(w, "Invalid access token")
			return
		}

		ctx := r.Context()

		var isRevoked bool
		revocationQuery := `SELECT EXISTS (SELECT 1 FROM revoked_tokens WHERE jti = $1)`
		err = m.db.QueryRow(ctx, revocationQuery, claims.ID).Scan(&isRevoked)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			m.errorLog.Error("failed to check revoked tokens", "error", err, "jti", claims.ID)
			utils.InternalServerError(w, err)
			return
		}

		if isRevoked {
			m.warnLog.Warn("revoked jti attempted access", "user_id", claims.UserID, "jti", claims.ID)
			utils.Unauthorized(w, "Session has been revoked")
			return
		}

		var currentVersion int
		var isActive bool
		userQuery := `SELECT token_version, is_active FROM users WHERE id = $1`

		err = m.db.QueryRow(ctx, userQuery, claims.UserID).Scan(&currentVersion, &isActive)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				m.warnLog.Warn("token subject user not found in database", "user_id", claims.UserID)
				utils.Unauthorized(w, "User no longer exists")
				return
			}
			m.errorLog.Error("database query failed in auth middleware", "error", err, "user_id", claims.UserID)
			utils.InternalServerError(w, err)
			return
		}

		if !isActive {
			m.warnLog.Warn("inactive user attempted access", "user_id", claims.UserID)
			utils.Forbidden(w, "Your account has been deactivated")
			return
		}

		if claims.TokenVersion != currentVersion {
			m.warnLog.Warn("outdated token version provided",
				"user_id", claims.UserID,
				"token_version", claims.TokenVersion,
				"current_version", currentVersion,
			)
			utils.Unauthorized(w, "Session expired due to a security update. Please log in again.")
			return
		}

		ctx = context.WithValue(ctx, UserContextKey, claims)
		ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(UserRoleKey).(string)
		if !ok || role != "admin" {
			m.warnLog.Warn("unauthorized access attempt to admin endpoint",
				"user_id", r.Context().Value(UserIDContextKey),
				"role", role,
				"path", r.URL.Path,
			)
			utils.Forbidden(w, "Administrator access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func GetUserClaims(ctx context.Context) (*utils.Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*utils.Claims)
	return claims, ok
}

func GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(UserIDContextKey).(string)
	return id, ok
}

func GetUserRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleKey).(string)
	return role, ok
}
