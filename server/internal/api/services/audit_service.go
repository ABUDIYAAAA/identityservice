package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"devclub.com/identity/internal/api/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// -----------------------------------------------------------------------------
// Audit Action Type Constants
// -----------------------------------------------------------------------------
const (
	// Authentication & Identity Events
	AuditActionUserLoginSuccess      = "USER_LOGIN_SUCCESS"
	AuditActionUserLoginFailed       = "USER_LOGIN_FAILED"
	AuditActionUserLoginLocked       = "USER_LOGIN_LOCKED"
	AuditActionUserPasswordChanged   = "USER_PASSWORD_CHANGED"
	AuditActionUserPasswordResetReq  = "USER_PASSWORD_RESET_REQUESTED"
	AuditActionUserPasswordResetDone = "USER_PASSWORD_RESET_COMPLETED"
	AuditActionUserLogout            = "USER_LOGOUT"
	AuditActionSessionRevoked        = "USER_SESSION_REVOKED"
	AuditActionSessionRevokeOthers   = "USER_SESSION_REVOKE_OTHERS"

	// Invitation & User Management Events
	AuditActionInviteCreated     = "INVITE_CREATED"
	AuditActionInviteAccepted    = "INVITE_ACCEPTED"
	AuditActionUserRoleUpdated   = "USER_ROLE_UPDATED"
	AuditActionUserStatusUpdated = "USER_STATUS_UPDATED"
	AuditActionUserDeleted       = "USER_DELETED"

	// Service Application Events
	AuditActionServiceCreated         = "SERVICE_CREATED"
	AuditActionServiceStatusUpdated   = "SERVICE_STATUS_UPDATED"
	AuditActionServiceDeleted         = "SERVICE_DELETED"
	AuditActionServiceSecretGenerated = "SERVICE_SECRET_GENERATED"
	AuditActionServiceSecretRevoked   = "SERVICE_SECRET_REVOKED"
	AuditActionServiceUserAssigned    = "SERVICE_USER_ASSIGNED"
	AuditActionServiceUserRevoked     = "SERVICE_USER_REVOKED"
	AuditActionServicePermGranted     = "SERVICE_PERMISSION_GRANTED"
	AuditActionServicePermRevoked     = "SERVICE_PERMISSION_REVOKED"
)

const (
	RedisAuditQueueKey = "audit:events:queue"
	BatchSize          = 50
	FlushInterval      = 2 * time.Second
)

type AuditClient interface {
	LogEvent(ctx context.Context, event models.AuditLogEvent)
	ListAuditLogs(ctx context.Context, actionType, actorID, serviceID string, limit, offset int) ([]models.AuditLogEvent, int64, error)
	StartWorker(ctx context.Context)
	Stop()
}

type redisAuditClient struct {
	rdb      *redis.Client
	db       *pgxpool.Pool
	errorLog *slog.Logger
	infoLog  *slog.Logger
	wg       sync.WaitGroup
	cancel   context.CancelFunc
}

func NewAuditClient(rdb *redis.Client, db *pgxpool.Pool, infoLog, errorLog *slog.Logger) AuditClient {
	return &redisAuditClient{
		rdb:      rdb,
		db:       db,
		infoLog:  infoLog,
		errorLog: errorLog,
	}
}

// LogEvent pushes an event asynchronously to Redis queue (fire-and-forget).
func (c *redisAuditClient) LogEvent(ctx context.Context, event models.AuditLogEvent) {
	if c.rdb == nil {
		return
	}

	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	if event.ActorType == "" {
		event.ActorType = "user"
	}
	if event.Status == "" {
		event.Status = "success"
	}

	data, err := json.Marshal(event)
	if err != nil {
		c.errorLog.Error("failed to marshal audit event", "error", err, "action", event.ActionType)
		return
	}

	// Non-blocking fire-and-forget push
	go func() {
		pushCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := c.rdb.RPush(pushCtx, RedisAuditQueueKey, data).Err(); err != nil {
			c.errorLog.Error("failed to push audit event to redis", "error", err, "action", event.ActionType)
		}
	}()
}

// StartWorker starts the background batch processing loop.
func (c *redisAuditClient) StartWorker(parentCtx context.Context) {
	if c.rdb == nil || c.db == nil {
		c.infoLog.Info("audit worker disabled (redis or db unavailable)")
		return
	}

	workerCtx, cancel := context.WithCancel(parentCtx)
	c.cancel = cancel
	c.wg.Add(1)

	go func() {
		defer c.wg.Done()
		c.infoLog.Info("starting async audit log batch worker")
		c.runBatchLoop(workerCtx)
	}()
}

func (c *redisAuditClient) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
}

func (c *redisAuditClient) runBatchLoop(ctx context.Context) {
	ticker := time.NewTicker(FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.flushQueue(context.Background())
			return
		case <-ticker.C:
			c.flushQueue(ctx)
		}
	}
}

func (c *redisAuditClient) flushQueue(ctx context.Context) {
	var events []models.AuditLogEvent

	for i := 0; i < BatchSize; i++ {
		popCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		res, err := c.rdb.LPop(popCtx, RedisAuditQueueKey).Result()
		cancel()

		if err == redis.Nil || err != nil {
			break
		}

		var ev models.AuditLogEvent
		if err := json.Unmarshal([]byte(res), &ev); err == nil {
			events = append(events, ev)
		}
	}

	if len(events) == 0 {
		return
	}

	if err := c.insertBatchDB(ctx, events); err != nil {
		c.errorLog.Error("failed to insert audit events batch into db", "error", err, "count", len(events))
	} else {
		c.infoLog.Info("flushed audit events batch to db", "count", len(events))
	}
}

func (c *redisAuditClient) insertBatchDB(ctx context.Context, events []models.AuditLogEvent) error {
	valueStrings := make([]string, 0, len(events))
	valueArgs := make([]any, 0, len(events)*11)

	for i, ev := range events {
		base := i * 11
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10, base+11))

		var beforeJSON, afterJSON any
		if ev.BeforeState != nil {
			if b, err := json.Marshal(ev.BeforeState); err == nil {
				beforeJSON = string(b)
			}
		}
		if ev.AfterState != nil {
			if a, err := json.Marshal(ev.AfterState); err == nil {
				afterJSON = string(a)
			}
		}

		var actorID *string
		if ev.ActorID != "" {
			actorID = &ev.ActorID
		}

		var serviceID *string
		if ev.ServiceID != "" {
			serviceID = &ev.ServiceID
		}

		valueArgs = append(valueArgs,
			ev.ActionType,
			ev.ActorType,
			actorID,
			serviceID,
			beforeJSON,
			afterJSON,
			ev.Status,
			ev.ErrorMessage,
			ev.IPAddress,
			ev.UserAgent,
			ev.CreatedAt,
		)
	}

	stmt := fmt.Sprintf(`
		INSERT INTO audit_logs 
		(action_type, actor_type, actor_id, service_id, before_state, after_state, status, error_message, ip_address, user_agent, created_at)
		VALUES %s
	`, strings.Join(valueStrings, ","))

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.db.Exec(dbCtx, stmt, valueArgs...)
	return err
}

func (c *redisAuditClient) ListAuditLogs(ctx context.Context, actionType, actorID, serviceID string, limit, offset int) ([]models.AuditLogEvent, int64, error) {
	if c.db == nil {
		return nil, 0, nil
	}

	var total int64
	countQuery := `
		SELECT COUNT(*) FROM audit_logs
		WHERE ($1 = '' OR action_type = $1)
		  AND ($2 = '' OR actor_id = CASE WHEN $2 = '' THEN NULL ELSE $2::uuid END)
		  AND ($3 = '' OR service_id = CASE WHEN $3 = '' THEN NULL ELSE $3::uuid END)
	`
	if err := c.db.QueryRow(ctx, countQuery, actionType, actorID, serviceID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, action_type, actor_type, COALESCE(actor_id::text, ''), COALESCE(service_id::text, ''),
		       before_state, after_state, status, COALESCE(error_message, ''), COALESCE(ip_address, ''), COALESCE(user_agent, ''), created_at
		FROM audit_logs
		WHERE ($1 = '' OR action_type = $1)
		  AND ($2 = '' OR actor_id = CASE WHEN $2 = '' THEN NULL ELSE $2::uuid END)
		  AND ($3 = '' OR service_id = CASE WHEN $3 = '' THEN NULL ELSE $3::uuid END)
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`
	rows, err := c.db.Query(ctx, query, actionType, actorID, serviceID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []models.AuditLogEvent
	for rows.Next() {
		var ev models.AuditLogEvent
		var beforeJSON, afterJSON []byte
		if err := rows.Scan(
			&ev.ID, &ev.ActionType, &ev.ActorType, &ev.ActorID, &ev.ServiceID,
			&beforeJSON, &afterJSON, &ev.Status, &ev.ErrorMessage, &ev.IPAddress, &ev.UserAgent, &ev.CreatedAt,
		); err != nil {
			return nil, 0, err
		}

		if len(beforeJSON) > 0 {
			_ = json.Unmarshal(beforeJSON, &ev.BeforeState)
		}
		if len(afterJSON) > 0 {
			_ = json.Unmarshal(afterJSON, &ev.AfterState)
		}

		events = append(events, ev)
	}

	return events, total, nil
}
