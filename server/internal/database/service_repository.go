package database

import (
	"context"
	"errors"
	"time"

	"devclub.com/identity/internal/api/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrDuplicateService    = errors.New("service with this name or client_id already exists")
	ErrDuplicateSecretName = errors.New("secret with this name already exists for this service")
	ErrPermissionExists    = errors.New("permission link between services already exists")
)

type ServiceRepository interface {
	// Service CRUD & Status
	CreateService(ctx context.Context, name, description, clientID, createdBy string) (*models.Service, error)
	GetServiceByID(ctx context.Context, id string) (*models.Service, error)
	GetServiceByClientID(ctx context.Context, clientID string) (*models.Service, error)
	ListServices(ctx context.Context, limit, offset int) ([]models.Service, int64, error)
	ListAssignedServices(ctx context.Context, userID string, limit, offset int) ([]models.Service, int64, error)
	UpdateServiceStatus(ctx context.Context, serviceID string, isActive bool) error
	IncrementServiceTokenVersion(ctx context.Context, serviceID string) error
	DeleteService(ctx context.Context, serviceID string) error

	// User-Service Assignments
	AssignUserToService(ctx context.Context, userID, serviceID string) error
	RemoveUserFromService(ctx context.Context, userID, serviceID string) error
	IsUserAssignedToService(ctx context.Context, userID, serviceID string) (bool, error)

	// Client Secrets
	CreateServiceSecret(ctx context.Context, serviceID, name, prefix, secretHash string, expiresAt *time.Time) (*models.ServiceSecret, error)
	ListServiceSecrets(ctx context.Context, serviceID string) ([]models.ServiceSecret, error)
	DeleteServiceSecret(ctx context.Context, secretID, serviceID string) error
	AuthenticateClientSecret(ctx context.Context, clientID, secretHash string) (*models.Service, error)

	// Service-to-Service Permissions
	AddServicePermission(ctx context.Context, sourceServiceID, targetServiceID string) error
	RemoveServicePermission(ctx context.Context, sourceServiceID, targetServiceID string) error
	HasPermission(ctx context.Context, sourceServiceID, targetServiceID string) (bool, error)
	ListAllowedTargetServices(ctx context.Context, sourceServiceID string) ([]models.Service, error)
}

type PostgresServiceRepository struct {
	db *pgxpool.Pool
}

func NewServiceRepository(db *pgxpool.Pool) ServiceRepository {
	return &PostgresServiceRepository{db: db}
}

// =========================================================================
// 1. Service CRUD & Administration
// =========================================================================

func (r *PostgresServiceRepository) CreateService(ctx context.Context, name, description, clientID, createdBy string) (*models.Service, error) {
	query := `
		INSERT INTO services (name, description, client_id, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, COALESCE(description, ''), client_id, is_active, token_version, created_by, created_at, updated_at
	`
	var s models.Service
	err := r.db.QueryRow(ctx, query, name, description, clientID, createdBy).Scan(
		&s.ID, &s.Name, &s.Description, &s.ClientID, &s.IsActive, &s.TokenVersion, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateService
		}
		return nil, err
	}
	return &s, nil
}

func (r *PostgresServiceRepository) GetServiceByID(ctx context.Context, id string) (*models.Service, error) {
	query := `
		SELECT id, name, COALESCE(description, ''), client_id, is_active, token_version, created_by, created_at, updated_at
		FROM services
		WHERE id = $1
	`
	var s models.Service
	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.Description, &s.ClientID, &s.IsActive, &s.TokenVersion, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *PostgresServiceRepository) GetServiceByClientID(ctx context.Context, clientID string) (*models.Service, error) {
	query := `
		SELECT id, name, COALESCE(description, ''), client_id, is_active, token_version, created_by, created_at, updated_at
		FROM services
		WHERE client_id = $1
	`
	var s models.Service
	err := r.db.QueryRow(ctx, query, clientID).Scan(
		&s.ID, &s.Name, &s.Description, &s.ClientID, &s.IsActive, &s.TokenVersion, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *PostgresServiceRepository) ListServices(ctx context.Context, limit, offset int) ([]models.Service, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM services`
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, name, COALESCE(description, ''), client_id, is_active, token_version, created_by, created_at, updated_at
		FROM services
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var services []models.Service
	for rows.Next() {
		var s models.Service
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.ClientID, &s.IsActive, &s.TokenVersion, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		services = append(services, s)
	}

	return services, total, nil
}

func (r *PostgresServiceRepository) ListAssignedServices(ctx context.Context, userID string, limit, offset int) ([]models.Service, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM user_services WHERE user_id = $1`
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT s.id, s.name, COALESCE(s.description, ''), s.client_id, s.is_active, s.token_version, s.created_by, s.created_at, s.updated_at
		FROM services s
		INNER JOIN user_services us ON us.service_id = s.id
		WHERE us.user_id = $1
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var services []models.Service
	for rows.Next() {
		var s models.Service
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.ClientID, &s.IsActive, &s.TokenVersion, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		services = append(services, s)
	}

	return services, total, nil
}

func (r *PostgresServiceRepository) UpdateServiceStatus(ctx context.Context, serviceID string, isActive bool) error {
	query := `
		UPDATE services 
		SET is_active = $1, token_version = token_version + 1, updated_at = NOW() 
		WHERE id = $2
	`
	ct, err := r.db.Exec(ctx, query, isActive, serviceID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresServiceRepository) IncrementServiceTokenVersion(ctx context.Context, serviceID string) error {
	query := `UPDATE services SET token_version = token_version + 1, updated_at = NOW() WHERE id = $1`
	ct, err := r.db.Exec(ctx, query, serviceID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresServiceRepository) DeleteService(ctx context.Context, serviceID string) error {
	query := `DELETE FROM services WHERE id = $1`
	ct, err := r.db.Exec(ctx, query, serviceID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// =========================================================================
// 2. User-Service Assignments
// =========================================================================

func (r *PostgresServiceRepository) AssignUserToService(ctx context.Context, userID, serviceID string) error {
	query := `
		INSERT INTO user_services (user_id, service_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, service_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, userID, serviceID)
	return err
}

func (r *PostgresServiceRepository) RemoveUserFromService(ctx context.Context, userID, serviceID string) error {
	query := `DELETE FROM user_services WHERE user_id = $1 AND service_id = $2`
	ct, err := r.db.Exec(ctx, query, userID, serviceID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresServiceRepository) IsUserAssignedToService(ctx context.Context, userID, serviceID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM user_services WHERE user_id = $1 AND service_id = $2)`
	err := r.db.QueryRow(ctx, query, userID, serviceID).Scan(&exists)
	return exists, err
}

// =========================================================================
// 3. Service Secrets
// =========================================================================

func (r *PostgresServiceRepository) CreateServiceSecret(ctx context.Context, serviceID, name, prefix, secretHash string, expiresAt *time.Time) (*models.ServiceSecret, error) {
	query := `
		INSERT INTO service_secrets (service_id, name, secret_prefix, secret_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, service_id, name, secret_prefix, last_used_at, expires_at, created_at
	`
	var s models.ServiceSecret
	err := r.db.QueryRow(ctx, query, serviceID, name, prefix, secretHash, expiresAt).Scan(
		&s.ID, &s.ServiceID, &s.Name, &s.SecretPrefix, &s.LastUsedAt, &s.ExpiresAt, &s.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateSecretName
		}
		return nil, err
	}

	return &s, nil
}

func (r *PostgresServiceRepository) ListServiceSecrets(ctx context.Context, serviceID string) ([]models.ServiceSecret, error) {
	query := `
		SELECT id, service_id, name, secret_prefix, last_used_at, expires_at, created_at
		FROM service_secrets
		WHERE service_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []models.ServiceSecret
	for rows.Next() {
		var s models.ServiceSecret
		if err := rows.Scan(&s.ID, &s.ServiceID, &s.Name, &s.SecretPrefix, &s.LastUsedAt, &s.ExpiresAt, &s.CreatedAt); err != nil {
			return nil, err
		}
		secrets = append(secrets, s)
	}

	return secrets, nil
}

func (r *PostgresServiceRepository) DeleteServiceSecret(ctx context.Context, secretID, serviceID string) error {
	query := `DELETE FROM service_secrets WHERE id = $1 AND service_id = $2`
	ct, err := r.db.Exec(ctx, query, secretID, serviceID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresServiceRepository) AuthenticateClientSecret(ctx context.Context, clientID, secretHash string) (*models.Service, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `
		SELECT s.id, s.name, COALESCE(s.description, ''), s.client_id, s.is_active, s.token_version, s.created_by, s.created_at, s.updated_at, sec.id
		FROM services s
		INNER JOIN service_secrets sec ON sec.service_id = s.id
		WHERE s.client_id = $1 
		  AND sec.secret_hash = $2 
		  AND (sec.expires_at IS NULL OR sec.expires_at > NOW())
		  AND s.is_active = TRUE
	`
	var s models.Service
	var secretID string
	err = tx.QueryRow(ctx, query, clientID, secretHash).Scan(
		&s.ID, &s.Name, &s.Description, &s.ClientID, &s.IsActive, &s.TokenVersion, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt, &secretID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Update last_used_at timestamp on authenticated secret
	_, _ = tx.Exec(ctx, `UPDATE service_secrets SET last_used_at = NOW() WHERE id = $1`, secretID)

	return &s, tx.Commit(ctx)
}

// =========================================================================
// 4. Service Permissions (ACL Matrix)
// =========================================================================

func (r *PostgresServiceRepository) AddServicePermission(ctx context.Context, sourceServiceID, targetServiceID string) error {
	query := `
		INSERT INTO service_permissions (source_service_id, target_service_id)
		VALUES ($1, $2)
	`
	_, err := r.db.Exec(ctx, query, sourceServiceID, targetServiceID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrPermissionExists
		}
		return err
	}
	return nil
}

func (r *PostgresServiceRepository) RemoveServicePermission(ctx context.Context, sourceServiceID, targetServiceID string) error {
	query := `DELETE FROM service_permissions WHERE source_service_id = $1 AND target_service_id = $2`
	ct, err := r.db.Exec(ctx, query, sourceServiceID, targetServiceID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresServiceRepository) HasPermission(ctx context.Context, sourceServiceID, targetServiceID string) (bool, error) {
	// A service always has authorization to call/verify against itself
	if sourceServiceID == targetServiceID {
		return true, nil
	}

	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM service_permissions WHERE source_service_id = $1 AND target_service_id = $2)`
	err := r.db.QueryRow(ctx, query, sourceServiceID, targetServiceID).Scan(&exists)
	return exists, err
}

func (r *PostgresServiceRepository) ListAllowedTargetServices(ctx context.Context, sourceServiceID string) ([]models.Service, error) {
	query := `
		SELECT s.id, s.name, COALESCE(s.description, ''), s.client_id, s.is_active, s.token_version, s.created_by, s.created_at, s.updated_at
		FROM services s
		INNER JOIN service_permissions sp ON sp.target_service_id = s.id
		WHERE sp.source_service_id = $1 AND s.is_active = TRUE
		ORDER BY s.name ASC
	`
	rows, err := r.db.Query(ctx, query, sourceServiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []models.Service
	for rows.Next() {
		var s models.Service
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.ClientID, &s.IsActive, &s.TokenVersion, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		targets = append(targets, s)
	}

	return targets, nil
}
