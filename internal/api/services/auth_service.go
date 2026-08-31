package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"devclub.com/identity/internal/api/config"
	"devclub.com/identity/internal/api/models"
	"devclub.com/identity/internal/database"
	"devclub.com/identity/internal/mailer"

	"devclub.com/identity/pkg/utils"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInviteNotFound     = errors.New("invite not found or expired")
	ErrSessionNotFound    = errors.New("session not found")
	ErrUnauthorized       = errors.New("unauthorized action")
)

type AuthService interface {
	Login(ctx context.Context, email, password, userAgent, ip string) (*utils.TokenPair, error)
	RefreshToken(ctx context.Context, rawRefreshToken, userAgent, ip string) (*utils.TokenPair, error)
	Logout(ctx context.Context, claims *utils.Claims, rawRefreshToken string) error

	// Invites & Password Resets
	CreateInvite(ctx context.Context, email, role, adminID string) (*models.UserInvitation, error)
	GetInviteDetails(ctx context.Context, token string) (*models.UserInvitation, error)
	AcceptInvite(ctx context.Context, token, password string) (*models.User, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error

	// Sessions & Revocation
	ListSessions(ctx context.Context, userID, currentTokenHash string) ([]models.RefreshTokenSession, error)
	RevokeSession(ctx context.Context, userID, sessionID string) error
	RevokeOtherSessions(ctx context.Context, userID, currentTokenHash string) error
	RevokeAllSessions(ctx context.Context, userID string) error
}

type authService struct {
	repo       database.AuthRepository
	jwtManager *utils.JWTManager
	tokenCache database.TokenCache
	mailer     *mailer.Mailer
	cfg        *config.Config
	infoLog    *slog.Logger
	warnLog    *slog.Logger
	errorLog   *slog.Logger
}

func NewAuthService(
	repo database.AuthRepository,
	jwtManager *utils.JWTManager,
	tokenCache database.TokenCache,
	mailer *mailer.Mailer,
	cfg *config.Config,
	infoLog, warnLog, errorLog *slog.Logger,
) AuthService {
	return &authService{
		repo:       repo,
		jwtManager: jwtManager,
		tokenCache: tokenCache,
		mailer:     mailer,
		cfg:        cfg,
		infoLog:    infoLog,
		warnLog:    warnLog,
		errorLog:   errorLog,
	}
}

// -----------------------------------------------------------------------------
// Authentication
// -----------------------------------------------------------------------------

func (s *authService) Login(ctx context.Context, email, password, userAgent, ip string) (*utils.TokenPair, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	match, err := utils.CheckPassword(password, user.PasswordHash)
	if err != nil || !match {
		return nil, ErrInvalidCredentials
	}

	// Generate access and refresh tokens
	tokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.Role, user.TokenVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	// Store session in DB
	_, err = s.repo.CreateSession(ctx, user.ID, tokens.RefreshTokenHash, userAgent, ip, tokens.RefreshExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	if s.tokenCache != nil {
		_ = s.tokenCache.CacheUserStatus(ctx, user.ID, user.TokenVersion, user.IsActive, s.cfg.JWTAccessTTL)
	}

	return tokens, nil
}

func (s *authService) RefreshToken(ctx context.Context, rawRefreshToken, userAgent, ip string) (*utils.TokenPair, error) {
	claims, err := s.jwtManager.VerifyRefreshToken(rawRefreshToken)
	if err != nil {
		return nil, ErrUnauthorized
	}

	// Compute hash of the submitted token to verify against database
	h := sha256.Sum256([]byte(rawRefreshToken))
	tokenHash := hex.EncodeToString(h[:])

	session, err := s.repo.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, ErrUnauthorized
		}
		return nil, err
	}

	if session.IsRevoked || session.ExpiresAt.Before(time.Now()) {
		return nil, ErrUnauthorized
	}

	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil || !user.IsActive || user.TokenVersion != claims.TokenVersion {
		return nil, ErrUnauthorized
	}

	// Invalidate previous refresh token (Refresh Token Rotation)
	_ = s.repo.RevokeSession(ctx, session.ID, user.ID)

	// Issue new token pair
	newTokens, err := s.jwtManager.GenerateTokenPair(user.ID, user.Role, user.TokenVersion)
	if err != nil {
		return nil, err
	}

	// Persist new session
	_, err = s.repo.CreateSession(ctx, user.ID, newTokens.RefreshTokenHash, userAgent, ip, newTokens.RefreshExpiresAt)
	if err != nil {
		return nil, err
	}

	if s.tokenCache != nil {
		_ = s.tokenCache.CacheUserStatus(ctx, user.ID, user.TokenVersion, user.IsActive, s.cfg.JWTAccessTTL)
	}

	return newTokens, nil
}

func (s *authService) Logout(ctx context.Context, claims *utils.Claims, rawRefreshToken string) error {
	// Granularly revoke current access token by its JTI
	if claims != nil && claims.ID != "" {
		_ = s.repo.RevokeAccessToken(ctx, claims.ID, claims.UserID, claims.ExpiresAt.Time)
		if s.tokenCache != nil {
			_ = s.tokenCache.CacheRevokedJTI(ctx, claims.ID, s.cfg.JWTAccessTTL)
			_ = s.tokenCache.InvalidateAccessToken(ctx, claims.ID)
		}
	}

	// Invalidate corresponding refresh token
	if rawRefreshToken != "" {
		h := sha256.Sum256([]byte(rawRefreshToken))
		tokenHash := hex.EncodeToString(h[:])
		session, err := s.repo.GetSessionByTokenHash(ctx, tokenHash)
		if err == nil {
			_ = s.repo.RevokeSession(ctx, session.ID, claims.UserID)
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Invites & Password Recovery
// -----------------------------------------------------------------------------

func (s *authService) CreateInvite(ctx context.Context, email, role, adminID string) (*models.UserInvitation, error) {
	// Check if user already exists
	if _, err := s.repo.GetUserByEmail(ctx, email); err == nil {
		return nil, ErrUserAlreadyExists
	}

	// Generate secure random invite token
	rawTokenBytes := make([]byte, 32)
	if _, err := rand.Read(rawTokenBytes); err != nil {
		return nil, err
	}
	rawToken := hex.EncodeToString(rawTokenBytes)

	h := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(h[:])
	expiresAt := time.Now().Add(24 * time.Hour)

	invitation, err := s.repo.CreateInvitation(ctx, email, role, tokenHash, adminID, expiresAt)
	if err != nil {
		return nil, err
	}

	// Fire off asynchronous invite email
	s.mailer.SendUserInvite(email, role, rawToken, s.cfg.FrontendURL)

	return invitation, nil
}

func (s *authService) GetInviteDetails(ctx context.Context, token string) (*models.UserInvitation, error) {
	h := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(h[:])

	invite, err := s.repo.GetInvitationByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, ErrInviteNotFound
		}
		return nil, err
	}

	if invite.UsedAt != nil || invite.ExpiresAt.Before(time.Now()) {
		return nil, ErrInviteNotFound
	}

	return invite, nil
}

func (s *authService) AcceptInvite(ctx context.Context, token, password string) (*models.User, error) {
	invite, err := s.GetInviteDetails(ctx, token)
	if err != nil {
		return nil, err
	}

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	h := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(h[:])

	user, err := s.repo.CreateUserFromInvite(ctx, invite.Email, passwordHash, invite.Role, tokenHash)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		// Silent exit to prevent email enumeration
		return nil
	}

	// Generate reset token
	rawTokenBytes := make([]byte, 32)
	if _, err := rand.Read(rawTokenBytes); err != nil {
		return err
	}
	rawToken := hex.EncodeToString(rawTokenBytes)

	h := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(h[:])
	expiresAt := time.Now().Add(15 * time.Minute)

	// Reuse user_invitations or dedicated reset tokens table using system admin ID
	_, err = s.repo.CreateInvitation(ctx, user.Email, user.Role, tokenHash, user.ID, expiresAt)
	if err != nil {
		return err
	}

	s.mailer.SendPasswordReset(user.Email, rawToken, s.cfg.FrontendURL)
	return nil
}

func (s *authService) ResetPassword(ctx context.Context, token, newPassword string) error {
	h := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(h[:])

	invite, err := s.repo.GetInvitationByHash(ctx, tokenHash)
	if err != nil || invite.UsedAt != nil || invite.ExpiresAt.Before(time.Now()) {
		return ErrInviteNotFound
	}

	user, err := s.repo.GetUserByEmail(ctx, invite.Email)
	if err != nil {
		return ErrInvalidCredentials
	}

	newHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password and increment token_version (kills all previous sessions)
	err = s.repo.UpdateUserPassword(ctx, user.ID, newHash)
	if err == nil && s.tokenCache != nil {
		_ = s.tokenCache.InvalidateUserStatus(ctx, user.ID)
	}
	return err
}

func (s *authService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	match, err := utils.CheckPassword(oldPassword, user.PasswordHash)
	if err != nil || !match {
		return ErrInvalidCredentials
	}

	newHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	err = s.repo.UpdateUserPassword(ctx, userID, newHash)
	if err == nil && s.tokenCache != nil {
		_ = s.tokenCache.InvalidateUserStatus(ctx, userID)
	}
	return err
}

// -----------------------------------------------------------------------------
// Session Management
// -----------------------------------------------------------------------------

func (s *authService) ListSessions(ctx context.Context, userID, currentTokenHash string) ([]models.RefreshTokenSession, error) {
	return s.repo.ListUserSessions(ctx, userID)
}

func (s *authService) RevokeSession(ctx context.Context, userID, sessionID string) error {
	return s.repo.RevokeSession(ctx, sessionID, userID)
}

func (s *authService) RevokeOtherSessions(ctx context.Context, userID, currentTokenHash string) error {
	session, err := s.repo.GetSessionByTokenHash(ctx, currentTokenHash)
	if err != nil {
		return ErrSessionNotFound
	}

	return s.repo.RevokeAllUserSessionsExceptCurrent(ctx, userID, session.ID)
}

func (s *authService) RevokeAllSessions(ctx context.Context, userID string) error {
	err := s.repo.RevokeAllUserSessions(ctx, userID)
	if err == nil && s.tokenCache != nil {
		_ = s.tokenCache.InvalidateUserStatus(ctx, userID)
	}
	return err
}
