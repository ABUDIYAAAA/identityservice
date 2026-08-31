package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"devclub.com/identity/internal/api/dto"
	"devclub.com/identity/internal/api/middlewares"
	"devclub.com/identity/internal/api/services"
	"devclub.com/identity/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type SessionHandler struct {
	authService services.AuthService
}

func NewSessionHandler(authService services.AuthService) *SessionHandler {
	return &SessionHandler{authService: authService}
}

func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserID(r.Context())
	if !ok {
		utils.Unauthorized(w, "Unauthorized")
		return
	}

	var currentTokenHash string
	if cookie, err := r.Cookie(utils.RefreshCookieName); err == nil && cookie.Value != "" {
		hash := sha256.Sum256([]byte(cookie.Value))
		currentTokenHash = hex.EncodeToString(hash[:])
	}

	sessions, err := h.authService.ListSessions(r.Context(), userID, currentTokenHash)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	sessionResponses := make([]dto.SessionResponse, len(sessions))
	for i, s := range sessions {
		sessionResponses[i] = dto.SessionResponse{
			ID:        s.ID,
			UserAgent: s.UserAgent,
			IPAddress: s.IPAddress,
			IsCurrent: currentTokenHash != "" && s.TokenHash == currentTokenHash,
			CreatedAt: s.CreatedAt,
			ExpiresAt: s.ExpiresAt,
		}
	}

	utils.Success(w, "Active sessions retrieved", sessionResponses)
}

func (h *SessionHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	userID, _ := middlewares.GetUserID(r.Context())
	sessionID := chi.URLParam(r, "sessionId")

	if sessionID == "" {
		utils.BadRequest(w, "Session ID is required", nil)
		return
	}

	if err := h.authService.RevokeSession(r.Context(), userID, sessionID); err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, "Session revoked successfully", nil)
}

func (h *SessionHandler) RevokeOtherSessions(w http.ResponseWriter, r *http.Request) {
	userID, _ := middlewares.GetUserID(r.Context())

	cookie, err := r.Cookie(utils.RefreshCookieName)
	if err != nil || cookie.Value == "" {
		utils.BadRequest(w, "Could not identify current refresh session cookie", nil)
		return
	}

	hsh := sha256.Sum256([]byte(cookie.Value))
	currentTokenHash := hex.EncodeToString(hsh[:])

	if err := h.authService.RevokeOtherSessions(r.Context(), userID, currentTokenHash); err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, "All other sessions have been revoked", nil)
}

func (h *SessionHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	userID, _ := middlewares.GetUserID(r.Context())

	if err := h.authService.RevokeAllSessions(r.Context(), userID); err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, "All sessions revoked. Please log in again.", nil)
}
