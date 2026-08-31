package database

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"devclub.com/identity/internal/api/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound       = errors.New("record not found")
	ErrDuplicateEmail = errors.New("email already registered")
	ErrDuplicateToken = errors.New("token already exists")
)

type AuthRepository interface {
	// User Management
	CreateUserFromInvite(ctx context.Context, email, passwordHash, role, inviteTokenHash string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	ListUsers(ctx context.Context, limit, offset int) ([]models.User, int64, error)
	UpdateUserPassword(ctx context.Context, userID, newPasswordHash string) error
	UpdateUserRole(ctx context.Context, userID, role string) error
	UpdateUserStatus(ctx context.Context, userID string, isActive bool) error
	IncrementTokenVersion(ctx context.Context, userID string) error
	DeleteUser(ctx context.Context, userID string) error

	// Invitation Management
	CreateInvitation(ctx context.Context, email, role, tokenHash, invitedBy string, expiresAt time.Time) (*models.UserInvitation, error)
	GetInvitationByHash(ctx context.Context, tokenHash string) (*models.UserInvitation, error)

	// Session & Refresh Tokens
	CreateSession(ctx context.Context, userID, tokenHash, userAgent, ip string, expiresAt time.Time) (*models.RefreshTokenSession, error)
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshTokenSession, error)
	ListUserSessions(ctx context.Context, userID string) ([]models.RefreshTokenSession, error)
	RevokeSession(ctx context.Context, sessionID, userID string) error
	RevokeAllUserSessionsExceptCurrent(ctx context.Context, userID, currentSessionID string) error
	RevokeAllUserSessions(ctx context.Context, userID string) error

	// Granular JWT Revocation (JTI)
	RevokeAccessToken(ctx context.Context, jti, userID string, expiresAt time.Time) error
	IsTokenRevoked(ctx context.Context, jti string) (bool, error)
}

type PostgresAuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) AuthRepository {
	return &PostgresAuthRepository{db: db}
}

// =========================================================================
// User Queries
// =========================================================================

func (r *PostgresAuthRepository) CreateUserFromInvite(ctx context.Context, email, passwordHash, role, inviteTokenHash string) (*models.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Mark invite as used
	markInviteQuery := `
		UPDATE user_invitations 
		SET used_at = NOW() 
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()
		RETURNING id
	`
	var inviteID string
	err = tx.QueryRow(ctx, markInviteQuery, inviteTokenHash).Scan(&inviteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invitation is invalid, expired, or already used")
		}
		return nil, err
	}

	// 2. Insert User
	insertUserQuery := `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, email, password_hash, role, is_active, token_version, created_at, updated_at
	`
	var u models.User
	err = tx.QueryRow(ctx, insertUserQuery, email, passwordHash, role).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.TokenVersion, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateEmail
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *PostgresAuthRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, role, is_active, token_version, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var u models.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.TokenVersion, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresAuthRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, role, is_active, token_version, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var u models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.TokenVersion, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresAuthRepository) ListUsers(ctx context.Context, limit, offset int) ([]models.User, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM users`
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, email, password_hash, role, is_active, token_version, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.TokenVersion, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	return users, total, nil
}

func (r *PostgresAuthRepository) UpdateUserPassword(ctx context.Context, userID, newPasswordHash string) error {
	query := `
		UPDATE users 
		SET password_hash = $1, token_version = token_version + 1, updated_at = NOW() 
		WHERE id = $2
	`
	ct, err := r.db.Exec(ctx, query, newPasswordHash, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresAuthRepository) UpdateUserRole(ctx context.Context, userID, role string) error {
	query := `UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2`
	ct, err := r.db.Exec(ctx, query, role, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresAuthRepository) UpdateUserStatus(ctx context.Context, userID string, isActive bool) error {
	query := `
		UPDATE users 
		SET is_active = $1, token_version = token_version + 1, updated_at = NOW() 
		WHERE id = $2
	`
	ct, err := r.db.Exec(ctx, query, isActive, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresAuthRepository) IncrementTokenVersion(ctx context.Context, userID string) error {
	query := `UPDATE users SET token_version = token_version + 1, updated_at = NOW() WHERE id = $1`
	ct, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresAuthRepository) DeleteUser(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE id = $1`
	ct, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// =========================================================================
// Invitation Queries
// =========================================================================

func (r *PostgresAuthRepository) CreateInvitation(ctx context.Context, email, role, tokenHash, invitedBy string, expiresAt time.Time) (*models.UserInvitation, error) {
	query := `
		INSERT INTO user_invitations (email, role, token_hash, invited_by, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, role, token_hash, invited_by, expires_at, used_at, created_at
	`
	var inv models.UserInvitation
	err := r.db.QueryRow(ctx, query, email, role, tokenHash, invitedBy, expiresAt).Scan(
		&inv.ID, &inv.Email, &inv.Role, &inv.TokenHash, &inv.InvitedBy, &inv.ExpiresAt, &inv.UsedAt, &inv.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateToken
		}
		return nil, err
	}
	return &inv, nil
}

func (r *PostgresAuthRepository) GetInvitationByHash(ctx context.Context, tokenHash string) (*models.UserInvitation, error) {
	query := `
		SELECT id, email, role, token_hash, invited_by, expires_at, used_at, created_at
		FROM user_invitations
		WHERE token_hash = $1
	`
	var inv models.UserInvitation
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&inv.ID, &inv.Email, &inv.Role, &inv.TokenHash, &inv.InvitedBy, &inv.ExpiresAt, &inv.UsedAt, &inv.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &inv, nil
}

// =========================================================================
// Session & Token Invalidation Queries
// =========================================================================

func (r *PostgresAuthRepository) CreateSession(ctx context.Context, userID, tokenHash, userAgent, ip string, expiresAt time.Time) (*models.RefreshTokenSession, error) {
	// Strip port if present in ip string (e.g. "[::1]:52365" -> "::1")
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}

	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, NULLIF($4, '')::inet, $5)
		RETURNING id, user_id, token_hash, user_agent, COALESCE(ip_address::text, ''), is_revoked, expires_at, created_at
	`
	var s models.RefreshTokenSession
	err := r.db.QueryRow(ctx, query, userID, tokenHash, userAgent, ip, expiresAt).Scan(
		&s.ID, &s.UserID, &s.TokenHash, &s.UserAgent, &s.IPAddress, &s.IsRevoked, &s.ExpiresAt, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *PostgresAuthRepository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshTokenSession, error) {
	query := `
		SELECT id, user_id, token_hash, user_agent, COALESCE(ip_address::text, ''), is_revoked, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	var s models.RefreshTokenSession
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&s.ID, &s.UserID, &s.TokenHash, &s.UserAgent, &s.IPAddress, &s.IsRevoked, &s.ExpiresAt, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *PostgresAuthRepository) ListUserSessions(ctx context.Context, userID string) ([]models.RefreshTokenSession, error) {
	query := `
		SELECT id, user_id, token_hash, user_agent, COALESCE(ip_address::text, ''), is_revoked, expires_at, created_at
		FROM refresh_tokens
		WHERE user_id = $1 AND is_revoked = FALSE AND expires_at > NOW()
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.RefreshTokenSession
	for rows.Next() {
		var s models.RefreshTokenSession
		if err := rows.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.UserAgent, &s.IPAddress, &s.IsRevoked, &s.ExpiresAt, &s.CreatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *PostgresAuthRepository) RevokeSession(ctx context.Context, sessionID, userID string) error {
	query := `
		UPDATE refresh_tokens 
		SET is_revoked = TRUE 
		WHERE id = $1 AND user_id = $2
	`
	ct, err := r.db.Exec(ctx, query, sessionID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresAuthRepository) RevokeAllUserSessionsExceptCurrent(ctx context.Context, userID, currentSessionID string) error {
	query := `
		UPDATE refresh_tokens 
		SET is_revoked = TRUE 
		WHERE user_id = $1 AND id != $2 AND is_revoked = FALSE
	`
	_, err := r.db.Exec(ctx, query, userID, currentSessionID)
	return err
}

func (r *PostgresAuthRepository) RevokeAllUserSessions(ctx context.Context, userID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Invalidate refresh tokens
	_, err = tx.Exec(ctx, `UPDATE refresh_tokens SET is_revoked = TRUE WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	// Bump token version to kill active access tokens
	_, err = tx.Exec(ctx, `UPDATE users SET token_version = token_version + 1, updated_at = NOW() WHERE id = $1`, userID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresAuthRepository) RevokeAccessToken(ctx context.Context, jti, userID string, expiresAt time.Time) error {
	query := `
		INSERT INTO revoked_tokens (jti, user_id, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (jti) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, jti, userID, expiresAt)
	return err
}

func (r *PostgresAuthRepository) IsTokenRevoked(ctx context.Context, jti string) (bool, error) {
	var isRevoked bool
	query := `SELECT EXISTS (SELECT 1 FROM revoked_tokens WHERE jti = $1)`
	err := r.db.QueryRow(ctx, query, jti).Scan(&isRevoked)
	return isRevoked, err
}
