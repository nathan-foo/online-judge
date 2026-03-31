package router

import (
	"log"
	"net/http"
	"time"

	"github.com/nathan-foo/online-judge/gateway/internal/auth"
	"github.com/nathan-foo/online-judge/gateway/internal/config"
	"github.com/nathan-foo/online-judge/gateway/internal/proxy"
	"github.com/nathan-foo/online-judge/gateway/internal/ratelimit"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewMux(allowedOrigins []string, routes []config.RouteConfig,
	authn *auth.Middleware, rl *ratelimit.RateLimiter) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           3600,
	}))
	r.Use(middleware.Timeout(20 * time.Second))
	r.Use(rl.Global())

	for _, route := range routes {
		handler, err := proxy.NewProxy(route.ServiceUrl, route.Prefix)
		if err != nil {
			log.Fatalf("Failed to create proxy for route %s: %v", route.Prefix, err)
		}
		r.Route(route.Prefix, func(r chi.Router) {
			if route.RequireAuth {
				r.Use(authn.WithAuth)
			}
			if route.RateLimit > 0 {
				r.Use(rl.Route(route.RateLimit))
			}
			if route.MaxUploadSize > 0 {
				r.Use(UploadHandler(route.MaxUploadSize))
			}
			r.Mount("/", handler)
		})
	}

	return r
}

func UploadHandler(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
