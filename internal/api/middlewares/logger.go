package middlewares

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			reqID := middleware.GetReqID(r.Context())

			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			attrs := []any{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.status),
				slog.Int("bytes", ww.bytes),
				slog.Duration("duration", duration),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			}

			if reqID != "" {
				attrs = append(attrs, slog.String("request_id", reqID))
			}

			switch {
			case ww.status >= 500:
				logger.Error("server error", attrs...)
			case ww.status >= 400:
				logger.Warn("client error", attrs...)
			default:
				logger.Info("request completed", attrs...)
			}
		})
	}
}
