package ratelimit

import (
	"net"
	"net/http"
	"strconv"

	"github.com/clerk/clerk-sdk-go/v2"
)

type keyFunc func(r *http.Request) (string, error)

func ipKey(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	return "ip:" + ip, nil
}

func userKey(r *http.Request) (string, error) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		return ipKey(r)
	}
	return "user:" + claims.Subject, nil
}

func newMiddleware(l *SlidingWindowLimiter, f keyFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key, err := f(r)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			result, err := l.Allow(r.Context(), key)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(l.limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))

			if !result.Allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(l.window.Seconds())))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
