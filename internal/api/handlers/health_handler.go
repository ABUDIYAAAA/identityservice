package handlers

import (
	"context"
	"net/http"
	"time"

	"devclub.com/identity/pkg/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	pool      *pgxpool.Pool
	rdb       *redis.Client
	startTime time.Time
}

func NewHealthHandler(pool *pgxpool.Pool, rdb *redis.Client, startTime time.Time) *HealthHandler {
	return &HealthHandler{
		pool:      pool,
		rdb:       rdb,
		startTime: startTime,
	}
}

type DBHealth struct {
	Status        string `json:"status"`
	Message       string `json:"message,omitempty"`
	TotalConns    int32  `json:"total_conns"`
	IdleConns     int32  `json:"idle_conns"`
	AcquiredConns int32  `json:"acquired_conns"`
	MaxConns      int32  `json:"max_conns"`
}

type RedisHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type HealthResponse struct {
	Status        string         `json:"status"`
	Timestamp     time.Time      `json:"timestamp"`
	Uptime        string         `json:"uptime"`
	UptimeSeconds float64        `json:"uptime_seconds"`
	Checks        map[string]any `json:"checks"`
}

func (h *HealthHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	isDegraded := false

	// Database Health Check
	dbHealth := DBHealth{
		Status: "unhealthy",
	}
	if h.pool == nil {
		dbHealth.Message = "database pool is not initialized"
		isDegraded = true
	} else {
		if err := h.pool.Ping(ctx); err != nil {
			dbHealth.Message = err.Error()
			isDegraded = true
		} else {
			dbHealth.Status = "healthy"
			dbHealth.Message = "connected"
		}

		stat := h.pool.Stat()
		dbHealth.TotalConns = stat.TotalConns()
		dbHealth.IdleConns = stat.IdleConns()
		dbHealth.AcquiredConns = stat.AcquiredConns()
		dbHealth.MaxConns = stat.MaxConns()
	}

	// Redis Health Check
	redisHealth := RedisHealth{
		Status: "unhealthy",
	}
	if h.rdb == nil {
		redisHealth.Message = "redis client is not initialized"
		isDegraded = true
	} else {
		if err := h.rdb.Ping(ctx).Err(); err != nil {
			redisHealth.Message = err.Error()
			isDegraded = true
		} else {
			redisHealth.Status = "healthy"
			redisHealth.Message = "connected"
		}
	}

	overallStatus := "healthy"
	if isDegraded {
		overallStatus = "degraded"
	}

	uptime := time.Since(h.startTime)

	response := HealthResponse{
		Status:        overallStatus,
		Timestamp:     time.Now().UTC(),
		Uptime:        uptime.Round(time.Second).String(),
		UptimeSeconds: uptime.Seconds(),
		Checks: map[string]any{
			"database": dbHealth,
			"redis":    redisHealth,
		},
	}

	utils.JSON(w, http.StatusOK, response)
}
