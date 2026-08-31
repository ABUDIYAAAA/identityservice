package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"devclub.com/identity/internal/api/dto"
	"devclub.com/identity/internal/api/middlewares"
	"devclub.com/identity/internal/api/services"
	"devclub.com/identity/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type ServiceHandler struct {
	serviceService services.ServiceService
}

func NewServiceHandler(serviceService services.ServiceService) *ServiceHandler {
	return &ServiceHandler{
		serviceService: serviceService,
	}
}

// =========================================================================
// 1. Service Administration & Management
// =========================================================================

func (h *ServiceHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middlewares.GetUserID(r.Context())
	if !ok {
		utils.Unauthorized(w, "Unauthorized")
		return
	}

	var req dto.CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Malformed JSON body", nil)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.BadRequest(w, "Validation failed", utils.FormatErrors(err))
		return
	}

	svc, secretResp, err := h.serviceService.CreateService(r.Context(), req.Name, req.Description, adminID)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.Created(w, "Service created successfully", map[string]any{
		"service":         svc,
		"initial_secret":  secretResp.RawSecret,
		"secret_metadata": secretResp.Secret,
	})
}

func (h *ServiceHandler) GetService(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "id")
	userID, _ := middlewares.GetUserID(r.Context())
	userRole, _ := middlewares.GetUserRole(r.Context())

	svc, err := h.serviceService.GetService(r.Context(), serviceID, userID, userRole)
	if err != nil {
		if errors.Is(err, services.ErrServiceNotFound) {
			utils.NotFound(w, "Service not found")
			return
		}
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Forbidden(w, "You are not assigned to manage this service")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, "Service retrieved successfully", svc)
}

func (h *ServiceHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	userID, _ := middlewares.GetUserID(r.Context())
	userRole, _ := middlewares.GetUserRole(r.Context())

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	svcList, total, err := h.serviceService.ListServices(r.Context(), userID, userRole, limit, offset)
	if err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, "Services retrieved successfully", dto.PaginatedServicesResponse{
		Services:   svcList,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	})
}

func (h *ServiceHandler) UpdateServiceStatus(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "id")

	var req dto.UpdateServiceStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Malformed JSON body", nil)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.BadRequest(w, "Validation failed", utils.FormatErrors(err))
		return
	}

	if err := h.serviceService.UpdateServiceStatus(r.Context(), serviceID, *req.IsActive); err != nil {
		if errors.Is(err, services.ErrServiceNotFound) {
			utils.NotFound(w, "Service not found")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, "Service status updated successfully", nil)
}

func (h *ServiceHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	adminID, _ := middlewares.GetUserID(r.Context())
	serviceID := chi.URLParam(r, "id")

	if err := h.serviceService.DeleteService(r.Context(), serviceID, adminID); err != nil {
		if errors.Is(err, services.ErrServiceNotFound) {
			utils.NotFound(w, "Service not found")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	utils.NoContent(w)
}

// =========================================================================
// 2. User-Service Assignments
// =========================================================================

func (h *ServiceHandler) AssignUser(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "id")

	var req dto.AssignUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Malformed JSON body", nil)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.BadRequest(w, "Validation failed", utils.FormatErrors(err))
		return
	}

	if err := h.serviceService.AssignUser(r.Context(), req.UserID, serviceID); err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, "User assigned to service successfully", nil)
}

func (h *ServiceHandler) RemoveUser(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "id")
	targetUserID := chi.URLParam(r, "userId")

	if err := h.serviceService.RemoveUser(r.Context(), targetUserID, serviceID); err != nil {
		if errors.Is(err, services.ErrServiceNotFound) {
			utils.NotFound(w, "Assignment not found")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	utils.NoContent(w)
}

// =========================================================================
// 3. Multi-Secret Management
// =========================================================================

func (h *ServiceHandler) GenerateSecret(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "id")
	userID, _ := middlewares.GetUserID(r.Context())
	userRole, _ := middlewares.GetUserRole(r.Context())

	var req dto.CreateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Malformed JSON body", nil)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.BadRequest(w, "Validation failed", utils.FormatErrors(err))
		return
	}

	secretResp, err := h.serviceService.GenerateSecret(r.Context(), serviceID, req.Name, req.ExpiresAt, userID, userRole)
	if err != nil {
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Forbidden(w, "You are not authorized to generate secrets for this service")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	utils.Created(w, "Secret generated successfully. Copy it now, it will not be displayed again.", map[string]any{
		"raw_secret":      secretResp.RawSecret,
		"secret_metadata": secretResp.Secret,
	})
}

func (h *ServiceHandler) ListSecrets(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "id")
	userID, _ := middlewares.GetUserID(r.Context())
	userRole, _ := middlewares.GetUserRole(r.Context())

	secrets, err := h.serviceService.ListSecrets(r.Context(), serviceID, userID, userRole)
	if err != nil {
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Forbidden(w, "Access denied")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, "Service secrets retrieved", secrets)
}

func (h *ServiceHandler) DeleteSecret(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "id")
	secretID := chi.URLParam(r, "secretId")
	userID, _ := middlewares.GetUserID(r.Context())
	userRole, _ := middlewares.GetUserRole(r.Context())

	if err := h.serviceService.DeleteSecret(r.Context(), secretID, serviceID, userID, userRole); err != nil {
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Forbidden(w, "Access denied")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	utils.NoContent(w)
}

// =========================================================================
// 4. Permissions (Inter-service links)
// =========================================================================

func (h *ServiceHandler) AddPermission(w http.ResponseWriter, r *http.Request) {
	sourceServiceID := chi.URLParam(r, "id")

	var req dto.CreatePermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Malformed JSON body", nil)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.BadRequest(w, "Validation failed", utils.FormatErrors(err))
		return
	}

	if err := h.serviceService.AddPermission(r.Context(), sourceServiceID, req.TargetServiceID); err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.Created(w, "Service access link granted successfully", nil)
}

func (h *ServiceHandler) RemovePermission(w http.ResponseWriter, r *http.Request) {
	sourceServiceID := chi.URLParam(r, "id")
	targetServiceID := chi.URLParam(r, "targetId")

	if err := h.serviceService.RemovePermission(r.Context(), sourceServiceID, targetServiceID); err != nil {
		utils.InternalServerError(w, err)
		return
	}

	utils.NoContent(w)
}

func (h *ServiceHandler) ListAllowedTargets(w http.ResponseWriter, r *http.Request) {
	sourceServiceID := chi.URLParam(r, "id")
	userID, _ := middlewares.GetUserID(r.Context())
	userRole, _ := middlewares.GetUserRole(r.Context())

	targets, err := h.serviceService.ListAllowedTargets(r.Context(), sourceServiceID, userID, userRole)
	if err != nil {
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Forbidden(w, "Access denied")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, "Allowed target services retrieved", targets)
}

// =========================================================================
// 5. M2M Token Generation & Verification
// =========================================================================

func (h *ServiceHandler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	var req dto.ServiceTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Malformed JSON body", nil)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.BadRequest(w, "Validation failed", utils.FormatErrors(err))
		return
	}

	tokenResp, err := h.serviceService.GenerateServiceToken(r.Context(), req.ClientID, req.ClientSecret)
	if err != nil {
		if errors.Is(err, services.ErrInvalidClientSecret) {
			utils.Unauthorized(w, "Invalid client ID or client secret")
			return
		}
		if errors.Is(err, services.ErrServiceInactive) {
			utils.Forbidden(w, "Service has been deactivated")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, "Token generated successfully", tokenResp)
}

func (h *ServiceHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyServiceTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "Malformed JSON body", nil)
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.BadRequest(w, "Validation failed", utils.FormatErrors(err))
		return
	}

	result, err := h.serviceService.VerifyServiceToken(
		r.Context(),
		req.TargetClientID,
		req.TargetClientSecret,
		req.CallerToken,
	)
	if err != nil {
		if errors.Is(err, services.ErrInvalidClientSecret) {
			utils.Unauthorized(w, "Target service credentials are invalid")
			return
		}
		if errors.Is(err, services.ErrServiceInactive) {
			utils.Forbidden(w, "Target service is deactivated")
			return
		}
		if errors.Is(err, services.ErrInvalidServiceToken) {
			utils.Unauthorized(w, "Caller token is invalid, expired, or revoked")
			return
		}
		if errors.Is(err, services.ErrAccessNotAllowed) {
			utils.Forbidden(w, "Caller service does not have permission to access target service")
			return
		}
		utils.InternalServerError(w, err)
		return
	}

	utils.Success(w, "Service token verified successfully", result)
}
