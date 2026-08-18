package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"sanitation-operations/internal/ratelimit"
	"sanitation-operations/internal/security"
)

func RateLimit(limiter *ratelimit.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limiter == nil {
				next.ServeHTTP(w, r)
				return
			}
			decision := limiter.Allow(security.ClientIP(r))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(decision.Remaining))
			if !decision.Allowed {
				retry := int(decision.RetryAfter.Round(time.Second).Seconds())
				if retry < 1 {
					retry = 1
				}
				w.Header().Set("Retry-After", strconv.Itoa(retry))
				message := "request rate limit exceeded"
				http.Error(w, message, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	if duration <= 0 {
		duration = 30 * time.Second
	}
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, duration, `{"code":"timeout","message":"request timed out"}`)
	}
}
