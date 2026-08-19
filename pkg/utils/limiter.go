package utils

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

type RateLimiter struct {
	mu           sync.Mutex
	buckets      map[string]*bucket
	rate         float64
	burst        float64
	cleanupTimer time.Duration
}

func NewRateLimiter(rate float64, burst int, cleanupInterval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets:      make(map[string]*bucket),
		rate:         rate,
		burst:        float64(burst),
		cleanupTimer: cleanupInterval,
	}

	go rl.startCleanup()
	return rl
}

func (rl *RateLimiter) Allow(key string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.buckets[key]

	if !exists {
		rl.buckets[key] = &bucket{
			tokens:     rl.burst - 1,
			lastRefill: now,
		}
		return true, 0
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens = math.Min(rl.burst, b.tokens+(elapsed*rl.rate))
	b.lastRefill = now

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true, 0
	}

	missingTokens := 1.0 - b.tokens
	retryAfter := time.Duration(missingTokens/rl.rate*float64(time.Second)) + 100*time.Millisecond

	return false, retryAfter
}

// Background cleanup routine to prevent memory exhaustion
func (rl *RateLimiter) startCleanup() {
	ticker := time.NewTicker(rl.cleanupTimer)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, b := range rl.buckets {
			// If bucket is full and inactive for more than 5 minutes, purge it
			if now.Sub(b.lastRefill) > 5*time.Minute {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

// ExtractClientKey combines the real IP and the client-supplied Device ID.
func ExtractClientKey(r *http.Request) string {
	deviceID := strings.TrimSpace(r.Header.Get("X-Device-ID"))
	if deviceID == "" {
		deviceID = "anonymous_device"
	}

	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// Take the first IP if multiple are present in X-Forwarded-For
		ips := strings.Split(ip, ",")
		ip = strings.TrimSpace(ips[0])
	} else {
		ip = r.Header.Get("X-Real-IP")
	}

	if ip == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			ip = host
		} else {
			ip = r.RemoteAddr
		}
	}

	return fmt.Sprintf("%s:%s", ip, deviceID)
}
