package middlewares

import (
	"fmt"
	"net/http"

	"devclub.com/identity/pkg/utils"
)

func PerRouteRateLimit(limiter *utils.RateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientKey := utils.ExtractClientKey(r)

			allowed, retryAfter := limiter.Allow(clientKey)
			if !allowed {
				retrySeconds := int(retryAfter.Seconds())
				if retrySeconds < 1 {
					retrySeconds = 1
				}

				w.Header().Set("Retry-After", fmt.Sprintf("%d", retrySeconds))
				utils.Error(w, http.StatusTooManyRequests, "Too many requests. Please slow down.", map[string]any{
					"retry_after_seconds": retrySeconds,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
