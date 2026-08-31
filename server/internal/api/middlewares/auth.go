package middlewares

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"devclub.com/identity/internal/database"
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
	tokenCache database.TokenCache
	accessTTL  time.Duration
	warnLog    *slog.Logger
	errorLog   *slog.Logger
}

func NewAuthMiddleware(jwtManager *utils.JWTManager, db *pgxpool.Pool, tokenCache database.TokenCache, accessTTL time.Duration, warnLog, errorLog *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		db:         db,
		tokenCache: tokenCache,
		accessTTL:  accessTTL,
		warnLog:    warnLog,
		errorLog:   errorLog,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenStr string

		if cookie, err := r.Cookie(utils.AccessCookieName); err == nil && cookie.Value != "" {
			tokenStr = cookie.Value
		} else {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
					tokenStr = parts[1]
				}
			}
		}

		if tokenStr == "" {
			utils.Unauthorized(w, "Authentication cookie or token required")
			return
		}

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

		// 1. Check if token JTI is revoked (Redis fast path)
		if m.tokenCache != nil {
			if revoked, err := m.tokenCache.IsJTIRevoked(ctx, claims.ID); err == nil && revoked {
				m.warnLog.Warn("revoked jti attempted access (cached)", "user_id", claims.UserID, "jti", claims.ID)
				utils.Unauthorized(w, "Session has been revoked")
				return
			}
		}

		var isRevoked bool
		if m.db != nil {
			revocationQuery := `SELECT EXISTS (SELECT 1 FROM revoked_tokens WHERE jti = $1)`
			err = m.db.QueryRow(ctx, revocationQuery, claims.ID).Scan(&isRevoked)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				m.errorLog.Error("failed to check revoked tokens", "error", err, "jti", claims.ID)
				utils.InternalServerError(w, err)
				return
			}
		}

		if isRevoked {
			if m.tokenCache != nil {
				_ = m.tokenCache.CacheRevokedJTI(ctx, claims.ID, m.accessTTL)
			}
			m.warnLog.Warn("revoked jti attempted access", "user_id", claims.UserID, "jti", claims.ID)
			utils.Unauthorized(w, "Session has been revoked")
			return
		}

		// 2. Check user version & status (Redis fast path)
		var currentVersion int
		var isActive bool
		foundInCache := false

		if m.tokenCache != nil {
			if status, found, err := m.tokenCache.GetUserStatus(ctx, claims.UserID); err == nil && found {
				currentVersion = status.TokenVersion
				isActive = status.IsActive
				foundInCache = true
			}
		}

		if !foundInCache {
			if m.db == nil {
				utils.InternalServerError(w, errors.New("database unavailable and status not cached"))
				return
			}
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

			if m.tokenCache != nil {
				_ = m.tokenCache.CacheUserStatus(ctx, claims.UserID, currentVersion, isActive, m.accessTTL)
			}
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
