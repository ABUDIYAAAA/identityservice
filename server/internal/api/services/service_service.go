package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"devclub.com/identity/internal/api/models"
	"devclub.com/identity/internal/database"
	"devclub.com/identity/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrServiceNotFound     = errors.New("service not found")
	ErrServiceInactive     = errors.New("service is deactivated")
	ErrInvalidClientSecret = errors.New("invalid client_id or client_secret")
	ErrPermissionDenied    = errors.New("insufficient permissions to manage this service")
	ErrAccessNotAllowed    = errors.New("caller service is not authorized to access the target service")
	ErrInvalidServiceToken = errors.New("invalid or expired service token")
)

type GeneratedSecretResponse struct {
	Secret    *models.ServiceSecret `json:"secret_metadata"`
	RawSecret string                `json:"raw_secret"` // Displayed strictly once upon creation
}

type ServiceTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type ServiceVerificationResult struct {
	Valid         bool            `json:"valid"`
	CallerService *models.Service `json:"caller_service"`
	TargetService *models.Service `json:"target_service"`
}

type ServiceService interface {
	// Service Administration
	CreateService(ctx context.Context, name, description, adminID string) (*models.Service, *GeneratedSecretResponse, error)
	GetService(ctx context.Context, serviceID, userID, userRole string) (*models.Service, error)
	ListServices(ctx context.Context, userID, userRole string, limit, offset int) ([]models.Service, int64, error)
	UpdateServiceStatus(ctx context.Context, serviceID string, isActive bool) error
	DeleteService(ctx context.Context, serviceID, adminID string) error

	// User-Service Assignment
	AssignUser(ctx context.Context, userID, serviceID string) error
	RemoveUser(ctx context.Context, userID, serviceID string) error

	// Secret Management
	GenerateSecret(ctx context.Context, serviceID, name string, expiresAt *time.Time, userID, userRole string) (*GeneratedSecretResponse, error)
	ListSecrets(ctx context.Context, serviceID, userID, userRole string) ([]models.ServiceSecret, error)
	DeleteSecret(ctx context.Context, secretID, serviceID, userID, userRole string) error

	// Service-to-Service Link Permissions
	AddPermission(ctx context.Context, sourceServiceID, targetServiceID string) error
	RemovePermission(ctx context.Context, sourceServiceID, targetServiceID string) error
	ListAllowedTargets(ctx context.Context, sourceServiceID, userID, userRole string) ([]models.Service, error)

	// M2M Authentication & Verification
	GenerateServiceToken(ctx context.Context, clientID, rawSecret string) (*ServiceTokenResponse, error)
	VerifyServiceToken(ctx context.Context, targetClientID, targetSecret, rawCallerToken string) (*ServiceVerificationResult, error)
}

type serviceService struct {
	repo        database.ServiceRepository
	tokenCache  database.TokenCache
	auditClient AuditClient
	jwtSecret   []byte
	tokenTTL    time.Duration
	infoLog     *slog.Logger
	warnLog     *slog.Logger
	errorLog    *slog.Logger
}

func NewServiceService(
	repo database.ServiceRepository,
	tokenCache database.TokenCache,
	auditClient AuditClient,
	jwtAccessSecret string,
	tokenTTL time.Duration,
	infoLog, warnLog, errorLog *slog.Logger,
) ServiceService {
	return &serviceService{
		repo:        repo,
		tokenCache:  tokenCache,
		auditClient: auditClient,
		jwtSecret:   []byte(jwtAccessSecret),
		tokenTTL:    tokenTTL,
		infoLog:     infoLog,
		warnLog:     warnLog,
		errorLog:    errorLog,
	}
}

// -----------------------------------------------------------------------------
// Access Control Checks
// -----------------------------------------------------------------------------

func (s *serviceService) ensureAccess(ctx context.Context, serviceID, userID, userRole string) error {
	if userRole == "admin" {
		return nil
	}

	assigned, err := s.repo.IsUserAssignedToService(ctx, userID, serviceID)
	if err != nil {
		return err
	}
	if !assigned {
		return ErrPermissionDenied
	}
	return nil
}

// -----------------------------------------------------------------------------
// 1. Service Administration
// -----------------------------------------------------------------------------

func (s *serviceService) CreateService(ctx context.Context, name, description, adminID string) (*models.Service, *GeneratedSecretResponse, error) {
	clientID, err := utils.GenerateClientID()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate client_id: %w", err)
	}

	rawSecret, prefix, secretHash, err := utils.GenerateClientSecret()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate client secret: %w", err)
	}

	svc, secret, err := s.repo.CreateServiceWithInitialSecret(ctx, name, description, clientID, adminID, "Primary Secret", prefix, secretHash, nil)
	if err != nil {
		return nil, nil, err
	}

	if s.auditClient != nil {
		s.auditClient.LogEvent(ctx, models.AuditLogEvent{
			ActionType: AuditActionServiceCreated,
			ActorType:  "user",
			ActorID:    adminID,
			ServiceID:  svc.ID,
			AfterState: map[string]any{"name": name, "client_id": clientID},
		})
	}

	return svc, &GeneratedSecretResponse{
		Secret:    secret,
		RawSecret: rawSecret,
	}, nil
}

func (s *serviceService) GetService(ctx context.Context, serviceID, userID, userRole string) (*models.Service, error) {
	if err := s.ensureAccess(ctx, serviceID, userID, userRole); err != nil {
		return nil, err
	}

	svc, err := s.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}

	return svc, nil
}

func (s *serviceService) ListServices(ctx context.Context, userID, userRole string, limit, offset int) ([]models.Service, int64, error) {
	if userRole == "admin" {
		return s.repo.ListServices(ctx, limit, offset)
	}
	return s.repo.ListAssignedServices(ctx, userID, limit, offset)
}

func (s *serviceService) UpdateServiceStatus(ctx context.Context, serviceID string, isActive bool) error {
	beforeSvc, _ := s.repo.GetServiceByID(ctx, serviceID)

	err := s.repo.UpdateServiceStatus(ctx, serviceID, isActive)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return ErrServiceNotFound
		}
		return err
	}
	if s.tokenCache != nil {
		_ = s.tokenCache.InvalidateServiceStatus(ctx, serviceID)
	}

	if s.auditClient != nil {
		var beforeState any
		if beforeSvc != nil {
			beforeState = map[string]any{"is_active": beforeSvc.IsActive}
		}
		s.auditClient.LogEvent(ctx, models.AuditLogEvent{
			ActionType:  AuditActionServiceStatusUpdated,
			ServiceID:   serviceID,
			BeforeState: beforeState,
			AfterState:  map[string]any{"is_active": isActive},
		})
	}

	return nil
}

func (s *serviceService) DeleteService(ctx context.Context, serviceID, adminID string) error {
	beforeSvc, _ := s.repo.GetServiceByID(ctx, serviceID)

	err := s.repo.DeleteService(ctx, serviceID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return ErrServiceNotFound
		}
		return err
	}
	if s.tokenCache != nil {
		_ = s.tokenCache.InvalidateServiceStatus(ctx, serviceID)
	}

	if s.auditClient != nil {
		var beforeState any
		if beforeSvc != nil {
			beforeState = map[string]any{"service_id": serviceID, "name": beforeSvc.Name, "client_id": beforeSvc.ClientID}
		} else {
			beforeState = map[string]any{"service_id": serviceID}
		}
		s.auditClient.LogEvent(ctx, models.AuditLogEvent{
			ActionType:  AuditActionServiceDeleted,
			ActorType:   "user",
			ActorID:     adminID,
			ServiceID:   "",
			BeforeState: beforeState,
		})
	}

	return nil
}

// -----------------------------------------------------------------------------
// 2. User-Service Assignments
// -----------------------------------------------------------------------------

func (s *serviceService) AssignUser(ctx context.Context, userID, serviceID string) error {
	err := s.repo.AssignUserToService(ctx, userID, serviceID)
	if err == nil && s.auditClient != nil {
		s.auditClient.LogEvent(ctx, models.AuditLogEvent{
			ActionType: AuditActionServiceUserAssigned,
			ServiceID:  serviceID,
			AfterState: map[string]any{"assigned_user_id": userID},
		})
	}
	return err
}

func (s *serviceService) RemoveUser(ctx context.Context, userID, serviceID string) error {
	err := s.repo.RemoveUserFromService(ctx, userID, serviceID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return ErrServiceNotFound
		}
		return err
	}
	if s.auditClient != nil {
		s.auditClient.LogEvent(ctx, models.AuditLogEvent{
			ActionType:  AuditActionServiceUserRevoked,
			ServiceID:   serviceID,
			BeforeState: map[string]any{"revoked_user_id": userID},
		})
	}
	return nil
}

// -----------------------------------------------------------------------------
// 3. Multi-Secret Management
// -----------------------------------------------------------------------------

func (s *serviceService) GenerateSecret(ctx context.Context, serviceID, name string, expiresAt *time.Time, userID, userRole string) (*GeneratedSecretResponse, error) {
	if err := s.ensureAccess(ctx, serviceID, userID, userRole); err != nil {
		return nil, err
	}

	rawSecret, prefix, secretHash, err := utils.GenerateClientSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	secret, err := s.repo.CreateServiceSecret(ctx, serviceID, name, prefix, secretHash, expiresAt)
	if err != nil {
		return nil, err
	}

	if s.auditClient != nil {
		s.auditClient.LogEvent(ctx, models.AuditLogEvent{
			ActionType: AuditActionServiceSecretGenerated,
			ActorType:  "user",
			ActorID:    userID,
			ServiceID:  serviceID,
			AfterState: map[string]any{"secret_name": name, "prefix": prefix},
		})
	}

	return &GeneratedSecretResponse{
		Secret:    secret,
		RawSecret: rawSecret,
	}, nil
}

func (s *serviceService) ListSecrets(ctx context.Context, serviceID, userID, userRole string) ([]models.ServiceSecret, error) {
	if err := s.ensureAccess(ctx, serviceID, userID, userRole); err != nil {
		return nil, err
	}

	return s.repo.ListServiceSecrets(ctx, serviceID)
}

func (s *serviceService) DeleteSecret(ctx context.Context, secretID, serviceID, userID, userRole string) error {
	if err := s.ensureAccess(ctx, serviceID, userID, userRole); err != nil {
		return err
	}

	err := s.repo.DeleteServiceSecret(ctx, secretID, serviceID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return errors.New("secret not found")
		}
		return err
	}

	if s.auditClient != nil {
		s.auditClient.LogEvent(ctx, models.AuditLogEvent{
			ActionType:  AuditActionServiceSecretRevoked,
			ActorType:   "user",
			ActorID:     userID,
			ServiceID:   serviceID,
			BeforeState: map[string]any{"secret_id": secretID},
		})
	}

	return nil
}

// -----------------------------------------------------------------------------
// 4. Permissions (Inter-service ACL links)
// -----------------------------------------------------------------------------

func (s *serviceService) AddPermission(ctx context.Context, sourceServiceID, targetServiceID string) error {
	err := s.repo.AddServicePermission(ctx, sourceServiceID, targetServiceID)
	if err == nil {
		if s.tokenCache != nil {
			_ = s.tokenCache.InvalidateServicePermission(ctx, sourceServiceID, targetServiceID)
		}
		if s.auditClient != nil {
			s.auditClient.LogEvent(ctx, models.AuditLogEvent{
				ActionType: AuditActionServicePermGranted,
				ServiceID:  sourceServiceID,
				AfterState: map[string]any{"target_service_id": targetServiceID},
			})
		}
	}
	return err
}

func (s *serviceService) RemovePermission(ctx context.Context, sourceServiceID, targetServiceID string) error {
	err := s.repo.RemoveServicePermission(ctx, sourceServiceID, targetServiceID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return errors.New("permission link not found")
		}
		return err
	}
	if s.tokenCache != nil {
		_ = s.tokenCache.InvalidateServicePermission(ctx, sourceServiceID, targetServiceID)
	}
	if s.auditClient != nil {
		s.auditClient.LogEvent(ctx, models.AuditLogEvent{
			ActionType:  AuditActionServicePermRevoked,
			ServiceID:   sourceServiceID,
			BeforeState: map[string]any{"target_service_id": targetServiceID},
		})
	}
	return nil
}

func (s *serviceService) ListAllowedTargets(ctx context.Context, sourceServiceID, userID, userRole string) ([]models.Service, error) {
	if err := s.ensureAccess(ctx, sourceServiceID, userID, userRole); err != nil {
		return nil, err
	}

	return s.repo.ListAllowedTargetServices(ctx, sourceServiceID)
}

// -----------------------------------------------------------------------------
// 5. M2M Token Generation & Target Verification
// -----------------------------------------------------------------------------

type ServiceJWTClaims struct {
	ServiceID    string `json:"service_id"`
	ClientID     string `json:"client_id"`
	TokenVersion int    `json:"token_version"`
	jwt.RegisteredClaims
}

func (s *serviceService) GenerateServiceToken(ctx context.Context, clientID, rawSecret string) (*ServiceTokenResponse, error) {
	secretHash := utils.HashSHA256(rawSecret)

	svc, err := s.repo.AuthenticateClientSecret(ctx, clientID, secretHash)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, ErrInvalidClientSecret
		}
		return nil, err
	}

	if !svc.IsActive {
		return nil, ErrServiceInactive
	}

	now := time.Now()
	expiresAt := now.Add(s.tokenTTL)

	claims := &ServiceJWTClaims{
		ServiceID:    svc.ID,
		ClientID:     svc.ClientID,
		TokenVersion: svc.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   svc.ID,
			Issuer:    "devclub-identity",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign service token: %w", err)
	}

	if s.tokenCache != nil {
		_ = s.tokenCache.CacheServiceStatus(ctx, svc.ID, svc.TokenVersion, svc.IsActive, s.tokenTTL)
	}

	return &ServiceTokenResponse{
		AccessToken: signedToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.tokenTTL.Seconds()),
	}, nil
}

func (s *serviceService) VerifyServiceToken(ctx context.Context, targetClientID, targetSecret, rawCallerToken string) (*ServiceVerificationResult, error) {
	// 1. Authenticate the receiving/verifying Target Service
	targetHash := utils.HashSHA256(targetSecret)
	targetService, err := s.repo.AuthenticateClientSecret(ctx, targetClientID, targetHash)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, ErrInvalidClientSecret
		}
		return nil, err
	}

	if !targetService.IsActive {
		return nil, ErrServiceInactive
	}

	// 2. Parse and Validate the Calling Service's M2M Token
	claims := &ServiceJWTClaims{}
	token, err := jwt.ParseWithClaims(rawCallerToken, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidServiceToken
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidServiceToken
	}

	// 3. Real-time Caller State & Token Version Verification (Redis fast path)
	var callerService *models.Service
	var callerVersion int
	var callerActive bool
	foundStatusInCache := false

	if s.tokenCache != nil {
		if st, found, err := s.tokenCache.GetServiceStatus(ctx, claims.ServiceID); err == nil && found {
			callerVersion = st.TokenVersion
			callerActive = st.IsActive
			foundStatusInCache = true
		}
	}

	if foundStatusInCache {
		if !callerActive || callerVersion != claims.TokenVersion {
			return nil, ErrInvalidServiceToken
		}
		callerService = &models.Service{
			ID:           claims.ServiceID,
			ClientID:     claims.ClientID,
			TokenVersion: callerVersion,
			IsActive:     callerActive,
		}
	} else {
		cs, err := s.repo.GetServiceByID(ctx, claims.ServiceID)
		if err != nil {
			return nil, ErrInvalidServiceToken
		}
		if !cs.IsActive || cs.TokenVersion != claims.TokenVersion {
			return nil, ErrInvalidServiceToken
		}
		callerService = cs
		if s.tokenCache != nil {
			_ = s.tokenCache.CacheServiceStatus(ctx, cs.ID, cs.TokenVersion, cs.IsActive, s.tokenTTL)
		}
	}

	// 4. Verify ACL Link Permissions (Redis fast path)
	hasPerm := false
	foundPermInCache := false
	if s.tokenCache != nil {
		if allowed, cached, err := s.tokenCache.GetServicePermission(ctx, callerService.ID, targetService.ID); err == nil && cached {
			hasPerm = allowed
			foundPermInCache = true
		}
	}

	if !foundPermInCache {
		perm, err := s.repo.HasPermission(ctx, callerService.ID, targetService.ID)
		if err != nil {
			return nil, err
		}
		hasPerm = perm
		if s.tokenCache != nil {
			_ = s.tokenCache.CacheServicePermission(ctx, callerService.ID, targetService.ID, perm, s.tokenTTL)
		}
	}

	if !hasPerm {
		return nil, ErrAccessNotAllowed
	}

	return &ServiceVerificationResult{
		Valid:         true,
		CallerService: callerService,
		TargetService: targetService,
	}, nil
}
