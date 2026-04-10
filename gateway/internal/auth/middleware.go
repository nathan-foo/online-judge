package auth

import (
	"net/http"

	"github.com/nathan-foo/online-judge/gateway/internal/config"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
)

type Middleware struct {
	authorize func(http.Handler) http.Handler
}

func NewMiddleware(authConfig config.AuthConfig) *Middleware {
	clerk.SetKey(authConfig.ClerkSecretKey)
	return &Middleware{
		authorize: clerkhttp.WithHeaderAuthorization(),
	}
}

func (m *Middleware) WithAuth(next http.Handler) http.Handler {
	return m.authorize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := clerk.SessionClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-User-ID", claims.Subject)
		next.ServeHTTP(w, r)
	}))
}
