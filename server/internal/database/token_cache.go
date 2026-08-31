package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type CachedUserStatus struct {
	TokenVersion int  `json:"token_version"`
	IsActive     bool `json:"is_active"`
}

type CachedAccessToken struct {
	UserID       string `json:"user_id"`
	Role         string `json:"role"`
	TokenVersion int    `json:"token_version"`
}

type CachedServiceStatus struct {
	TokenVersion int  `json:"token_version"`
	IsActive     bool `json:"is_active"`
}

type TokenCache interface {
	CacheUserStatus(ctx context.Context, userID string, tokenVersion int, isActive bool, ttl time.Duration) error
	GetUserStatus(ctx context.Context, userID string) (*CachedUserStatus, bool, error)
	InvalidateUserStatus(ctx context.Context, userID string) error

	CacheRevokedJTI(ctx context.Context, jti string, ttl time.Duration) error
	IsJTIRevoked(ctx context.Context, jti string) (bool, error)

	CacheAccessToken(ctx context.Context, jti string, userID string, role string, tokenVersion int, ttl time.Duration) error
	GetAccessToken(ctx context.Context, jti string) (*CachedAccessToken, bool, error)
	InvalidateAccessToken(ctx context.Context, jti string) error

	// Service Token & ACL Caching
	CacheServiceStatus(ctx context.Context, serviceID string, tokenVersion int, isActive bool, ttl time.Duration) error
	GetServiceStatus(ctx context.Context, serviceID string) (*CachedServiceStatus, bool, error)
	InvalidateServiceStatus(ctx context.Context, serviceID string) error

	CacheServicePermission(ctx context.Context, sourceServiceID, targetServiceID string, allowed bool, ttl time.Duration) error
	GetServicePermission(ctx context.Context, sourceServiceID, targetServiceID string) (allowed bool, cached bool, err error)
	InvalidateServicePermission(ctx context.Context, sourceServiceID, targetServiceID string) error

	// Brute-Force Lockout
	IncrementFailedLogins(ctx context.Context, email string) (int64, error)
	IsAccountLocked(ctx context.Context, email string) (bool, time.Duration, error)
	LockAccount(ctx context.Context, email string, ttl time.Duration) error
	ResetFailedLogins(ctx context.Context, email string) error
}

type redisTokenCache struct {
	rdb     *redis.Client
	warnLog *slog.Logger
}

func NewTokenCache(rdb *redis.Client, warnLog *slog.Logger) TokenCache {
	return &redisTokenCache{
		rdb:     rdb,
		warnLog: warnLog,
	}
}

func (c *redisTokenCache) CacheUserStatus(ctx context.Context, userID string, tokenVersion int, isActive bool, ttl time.Duration) error {
	if c.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("user:status:%s", userID)
	data, err := json.Marshal(CachedUserStatus{
		TokenVersion: tokenVersion,
		IsActive:     isActive,
	})
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, ttl).Err()
}

func (c *redisTokenCache) GetUserStatus(ctx context.Context, userID string) (*CachedUserStatus, bool, error) {
	if c.rdb == nil {
		return nil, false, nil
	}
	key := fmt.Sprintf("user:status:%s", userID)
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	var status CachedUserStatus
	if err := json.Unmarshal([]byte(val), &status); err != nil {
		return nil, false, err
	}
	return &status, true, nil
}

func (c *redisTokenCache) InvalidateUserStatus(ctx context.Context, userID string) error {
	if c.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("user:status:%s", userID)
	return c.rdb.Del(ctx, key).Err()
}

func (c *redisTokenCache) CacheRevokedJTI(ctx context.Context, jti string, ttl time.Duration) error {
	if c.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("token:revoked:%s", jti)
	return c.rdb.Set(ctx, key, "1", ttl).Err()
}

func (c *redisTokenCache) IsJTIRevoked(ctx context.Context, jti string) (bool, error) {
	if c.rdb == nil {
		return false, nil
	}
	key := fmt.Sprintf("token:revoked:%s", jti)
	exists, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (c *redisTokenCache) CacheAccessToken(ctx context.Context, jti string, userID string, role string, tokenVersion int, ttl time.Duration) error {
	if c.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("token:access:%s", jti)
	data, err := json.Marshal(CachedAccessToken{
		UserID:       userID,
		Role:         role,
		TokenVersion: tokenVersion,
	})
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, ttl).Err()
}

func (c *redisTokenCache) GetAccessToken(ctx context.Context, jti string) (*CachedAccessToken, bool, error) {
	if c.rdb == nil {
		return nil, false, nil
	}
	key := fmt.Sprintf("token:access:%s", jti)
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	var token CachedAccessToken
	if err := json.Unmarshal([]byte(val), &token); err != nil {
		return nil, false, err
	}
	return &token, true, nil
}

func (c *redisTokenCache) InvalidateAccessToken(ctx context.Context, jti string) error {
	if c.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("token:access:%s", jti)
	return c.rdb.Del(ctx, key).Err()
}

func (c *redisTokenCache) CacheServiceStatus(ctx context.Context, serviceID string, tokenVersion int, isActive bool, ttl time.Duration) error {
	if c.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("service:status:%s", serviceID)
	data, err := json.Marshal(CachedServiceStatus{
		TokenVersion: tokenVersion,
		IsActive:     isActive,
	})
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, ttl).Err()
}

func (c *redisTokenCache) GetServiceStatus(ctx context.Context, serviceID string) (*CachedServiceStatus, bool, error) {
	if c.rdb == nil {
		return nil, false, nil
	}
	key := fmt.Sprintf("service:status:%s", serviceID)
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	var status CachedServiceStatus
	if err := json.Unmarshal([]byte(val), &status); err != nil {
		return nil, false, err
	}
	return &status, true, nil
}

func (c *redisTokenCache) InvalidateServiceStatus(ctx context.Context, serviceID string) error {
	if c.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("service:status:%s", serviceID)
	return c.rdb.Del(ctx, key).Err()
}

func (c *redisTokenCache) CacheServicePermission(ctx context.Context, sourceServiceID, targetServiceID string, allowed bool, ttl time.Duration) error {
	if c.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("service:perm:%s:%s", sourceServiceID, targetServiceID)
	val := "0"
	if allowed {
		val = "1"
	}
	return c.rdb.Set(ctx, key, val, ttl).Err()
}

func (c *redisTokenCache) GetServicePermission(ctx context.Context, sourceServiceID, targetServiceID string) (bool, bool, error) {
	if c.rdb == nil {
		return false, false, nil
	}
	key := fmt.Sprintf("service:perm:%s:%s", sourceServiceID, targetServiceID)
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, false, nil
	} else if err != nil {
		return false, false, err
	}
	return val == "1", true, nil
}

func (c *redisTokenCache) InvalidateServicePermission(ctx context.Context, sourceServiceID, targetServiceID string) error {
	if c.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("service:perm:%s:%s", sourceServiceID, targetServiceID)
	return c.rdb.Del(ctx, key).Err()
}

func (c *redisTokenCache) IncrementFailedLogins(ctx context.Context, email string) (int64, error) {
	if c.rdb == nil {
		return 0, nil
	}
	key := fmt.Sprintf("lockout:failed:%s", strings.ToLower(email))
	pipe := c.rdb.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 15*time.Minute)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

func (c *redisTokenCache) IsAccountLocked(ctx context.Context, email string) (bool, time.Duration, error) {
	if c.rdb == nil {
		return false, 0, nil
	}
	key := fmt.Sprintf("lockout:locked:%s", strings.ToLower(email))
	ttl, err := c.rdb.TTL(ctx, key).Result()
	if err == redis.Nil || ttl <= 0 {
		return false, 0, nil
	} else if err != nil {
		return false, 0, err
	}
	return true, ttl, nil
}

func (c *redisTokenCache) LockAccount(ctx context.Context, email string, ttl time.Duration) error {
	if c.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("lockout:locked:%s", strings.ToLower(email))
	return c.rdb.Set(ctx, key, "1", ttl).Err()
}

func (c *redisTokenCache) ResetFailedLogins(ctx context.Context, email string) error {
	if c.rdb == nil {
		return nil
	}
	keyFailed := fmt.Sprintf("lockout:failed:%s", strings.ToLower(email))
	keyLocked := fmt.Sprintf("lockout:locked:%s", strings.ToLower(email))
	return c.rdb.Del(ctx, keyFailed, keyLocked).Err()
}
