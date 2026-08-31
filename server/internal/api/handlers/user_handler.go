package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"devclub.com/identity/internal/api/dto"
	"devclub.com/identity/internal/api/middlewares"
	"devclub.com/identity/internal/api/models"
	"devclub.com/identity/internal/api/services"
	"devclub.com/identity/internal/database"
	"devclub.com/identity/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	authService services.AuthService
	repo        database.AuthRepository
	tokenCache  database.TokenCache
	auditClient services.AuditClient
}

func NewUserHandler(authService services.AuthService, repo database.AuthRepository, tokenCache database.TokenCache, auditClient services.AuditClient) *UserHandler {
	return &UserHandler{
		authService: authService,
		repo:        repo,
		tokenCache:  tokenCache,
		auditClient: auditClient,
	}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserID(r.Context())
	if !ok {
		utils.Unauthorized(w, "Unauthorized")
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			utils.NotFound(w, "User not found")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, "Profile retrieved successfully", dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
	})
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserID(r.Context())
	if !ok {
		utils.Unauthorized(w, "Unauthorized")
		return
	}

	var req dto.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Malformed JSON request body", nil)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.BadRequest(w, "Validation failed", utils.FormatErrors(err))
		return
	}

	err := h.authService.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			utils.BadRequest(w, "Current password does not match", nil)
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, "Password updated successfully. All other sessions have been invalidated.", nil)
}

// ---------------- Admin Actions ----------------

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	users, total, err := h.repo.ListUsers(r.Context(), limit, offset)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	userResponses := make([]dto.UserResponse, len(users))
	for i, u := range users {
		userResponses[i] = dto.UserResponse{
			ID:        u.ID,
			Email:     u.Email,
			Role:      u.Role,
			IsActive:  u.IsActive,
			CreatedAt: u.CreatedAt,
		}
	}

	utils.Success(w, "Users retrieved", dto.PaginatedUsersResponse{
		Users:      userResponses,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	})
}

func (h *UserHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	adminID, _ := middlewares.GetUserID(r.Context())

	var req dto.CreateInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Malformed JSON request body", nil)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.BadRequest(w, "Validation failed", utils.FormatErrors(err))
		return
	}

	inv, err := h.authService.CreateInvite(r.Context(), req.Email, req.Role, adminID)
	if err != nil {
		if errors.Is(err, services.ErrUserAlreadyExists) {
			utils.BadRequest(w, "A user with this email already exists", nil)
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	utils.Created(w, "Invitation sent successfully", inv)
}

func (s *UserHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	adminID, _ := middlewares.GetUserID(r.Context())
	userID := chi.URLParam(r, "id")

	var req dto.UpdateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Malformed JSON request body", nil)
		return
	}
	req.UserID = userID

	if err := utils.ValidateStruct(req); err != nil {
		utils.BadRequest(w, "Validation failed", utils.FormatErrors(err))
		return
	}

	beforeUser, _ := s.repo.GetUserByID(r.Context(), userID)

	if err := s.repo.UpdateUserRole(r.Context(), userID, req.Role); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			utils.NotFound(w, "User not found")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	if s.tokenCache != nil {
		_ = s.tokenCache.InvalidateUserStatus(r.Context(), userID)
	}

	if s.auditClient != nil {
		var beforeState any
		if beforeUser != nil {
			beforeState = map[string]any{"role": beforeUser.Role}
		}
		s.auditClient.LogEvent(r.Context(), models.AuditLogEvent{
			ActionType:  services.AuditActionUserRoleUpdated,
			ActorType:   "user",
			ActorID:     adminID,
			BeforeState: beforeState,
			AfterState:  map[string]any{"role": req.Role, "target_user_id": userID},
			IPAddress:   r.RemoteAddr,
			UserAgent:   r.UserAgent(),
		})
	}

	utils.Success(w, "User role updated successfully", nil)
}

func (s *UserHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	adminID, _ := middlewares.GetUserID(r.Context())
	userID := chi.URLParam(r, "id")

	var req dto.UpdateUserStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Malformed JSON request body", nil)
		return
	}
	req.UserID = userID

	if err := utils.ValidateStruct(req); err != nil {
		utils.BadRequest(w, "Validation failed", utils.FormatErrors(err))
		return
	}

	beforeUser, _ := s.repo.GetUserByID(r.Context(), userID)

	if err := s.repo.UpdateUserStatus(r.Context(), userID, *req.IsActive); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			utils.NotFound(w, "User not found")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	if s.tokenCache != nil {
		_ = s.tokenCache.InvalidateUserStatus(r.Context(), userID)
	}

	if s.auditClient != nil {
		var beforeState any
		if beforeUser != nil {
			beforeState = map[string]any{"is_active": beforeUser.IsActive}
		}
		s.auditClient.LogEvent(r.Context(), models.AuditLogEvent{
			ActionType:  services.AuditActionUserStatusUpdated,
			ActorType:   "user",
			ActorID:     adminID,
			BeforeState: beforeState,
			AfterState:  map[string]any{"is_active": *req.IsActive, "target_user_id": userID},
			IPAddress:   r.RemoteAddr,
			UserAgent:   r.UserAgent(),
		})
	}

	utils.Success(w, "User status updated successfully", nil)
}

func (s *UserHandler) RemoveUser(w http.ResponseWriter, r *http.Request) {
	adminID, _ := middlewares.GetUserID(r.Context())
	userID := chi.URLParam(r, "id")

	beforeUser, _ := s.repo.GetUserByID(r.Context(), userID)

	if err := s.repo.DeleteUser(r.Context(), userID); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			utils.NotFound(w, "User not found")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	if s.tokenCache != nil {
		_ = s.tokenCache.InvalidateUserStatus(r.Context(), userID)
	}

	if s.auditClient != nil {
		var beforeState any
		if beforeUser != nil {
			beforeState = map[string]any{"email": beforeUser.Email, "role": beforeUser.Role}
		}
		s.auditClient.LogEvent(r.Context(), models.AuditLogEvent{
			ActionType:  services.AuditActionUserDeleted,
			ActorType:   "user",
			ActorID:     adminID,
			BeforeState: beforeState,
			IPAddress:   r.RemoteAddr,
			UserAgent:   r.UserAgent(),
		})
	}

	utils.NoContent(w)
}
