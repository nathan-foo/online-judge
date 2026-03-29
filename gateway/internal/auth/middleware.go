package auth

import (
	"net/http"

	"github.com/nathan-foo/online-judge/gateway/internal/config"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
)

type Middleware struct{}

func NewMiddleware(authConfig config.AuthConfig) (*Middleware, error) {
	clerk.SetKey(authConfig.ClerkSecretKey)
	return &Middleware{}, nil
}

func (m *Middleware) WithAuth(next http.Handler) http.Handler {
	return clerkhttp.WithHeaderAuthorization()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := clerk.SessionClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}))
}
