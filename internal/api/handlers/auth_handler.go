package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"devclub.com/identity/internal/api/dto"
	"devclub.com/identity/internal/api/middlewares"
	"devclub.com/identity/internal/api/services"
	"devclub.com/identity/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	authService services.AuthService
	jwtManager  *utils.JWTManager
}

func NewAuthHandler(authService services.AuthService, jwtManager *utils.JWTManager) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtManager:  jwtManager,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Malformed JSON request body", nil)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.BadRequest(w, "Validation failed", utils.FormatErrors(err))
		return
	}

	tokens, err := h.authService.Login(r.Context(), req.Email, req.Password, r.UserAgent(), r.RemoteAddr)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			utils.Unauthorized(w, "Invalid email or password")
			return
		}
		if errors.Is(err, services.ErrUserInactive) {
			utils.Forbidden(w, "Account has been deactivated")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	// Set secure HTTP-only refresh token cookie
	h.jwtManager.SetRefreshCookie(w, tokens.RefreshToken, tokens.RefreshExpiresAt)

	utils.Success(w, "Login successful", dto.AuthTokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Until(tokens.AccessExpiresAt).Seconds()),
	})
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var rawRefreshToken string

	// Prefer cookie, fallback to JSON payload
	if cookie, err := r.Cookie(utils.RefreshCookieName); err == nil {
		rawRefreshToken = cookie.Value
	} else {
		var req dto.RefreshTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			rawRefreshToken = req.RefreshToken
		}
	}

	if rawRefreshToken == "" {
		utils.Unauthorized(w, "Refresh token required")
		return
	}

	tokens, err := h.authService.RefreshToken(r.Context(), rawRefreshToken, r.UserAgent(), r.RemoteAddr)
	if err != nil {
		h.jwtManager.ClearRefreshCookie(w)
		utils.Unauthorized(w, "Invalid or expired refresh token")
		return
	}

	h.jwtManager.SetRefreshCookie(w, tokens.RefreshToken, tokens.RefreshExpiresAt)

	utils.Success(w, "Token refreshed", dto.AuthTokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Until(tokens.AccessExpiresAt).Seconds()),
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims, _ := middlewares.GetUserClaims(r.Context())

	var rawRefreshToken string
	if cookie, err := r.Cookie(utils.RefreshCookieName); err == nil {
		rawRefreshToken = cookie.Value
	}

	_ = h.authService.Logout(r.Context(), claims, rawRefreshToken)
	h.jwtManager.ClearRefreshCookie(w)

	utils.Success(w, "Logged out successfully", nil)
}

func (h *AuthHandler) GetInviteDetails(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		utils.BadRequest(w, "Invite token is required", nil)
		return
	}

	invite, err := h.authService.GetInviteDetails(r.Context(), token)
	if err != nil {
		utils.NotFound(w, "Invitation is invalid or expired")
		return
	}

	utils.Success(w, "Invite details retrieved", dto.InviteDetailsResponse{
		Email:     invite.Email,
		Role:      invite.Role,
		InvitedBy: invite.InvitedBy,
		ExpiresAt: invite.ExpiresAt,
	})
}

func (h *AuthHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var req dto.AcceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Malformed JSON request body", nil)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.BadRequest(w, "Validation failed", utils.FormatErrors(err))
		return
	}

	user, err := h.authService.AcceptInvite(r.Context(), req.Token, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInviteNotFound) {
			utils.BadRequest(w, "Invitation is invalid, expired, or already used", nil)
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	utils.Created(w, "Account created successfully. Please login.", dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
	})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Malformed JSON request body", nil)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.BadRequest(w, "Validation failed", utils.FormatErrors(err))
		return
	}

	_ = h.authService.ForgotPassword(r.Context(), req.Email)
	utils.Success(w, "If an account exists with that email, a password reset link has been sent.", nil)
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Malformed JSON request body", nil)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.BadRequest(w, "Validation failed", utils.FormatErrors(err))
		return
	}

	err := h.authService.ResetPassword(r.Context(), req.Token, req.NewPassword)
	if err != nil {
		if errors.Is(err, services.ErrInviteNotFound) {
			utils.BadRequest(w, "Password reset token is invalid or has expired", nil)
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, "Password has been successfully reset. Please log in with your new password.", nil)
}
