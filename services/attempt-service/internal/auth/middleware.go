package auth

import (
	"context"
	"encoding/json"
	"net/http"
)

type contextKey struct{}

var userIDKey = contextKey{}

// WithUser requires the X-User-ID header set by the gateway and stores
// the user id on the request context.
func WithUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"detail": "Missing user identity"})
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserID(ctx context.Context) string {
	userID, _ := ctx.Value(userIDKey).(string)
	return userID
}
