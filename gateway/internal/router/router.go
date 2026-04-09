package router

import (
	"net/http"
	"time"

	"github.com/nathan-foo/online-judge/gateway/internal/auth"
	"github.com/nathan-foo/online-judge/gateway/internal/config"
	"github.com/nathan-foo/online-judge/gateway/internal/health"
	"github.com/nathan-foo/online-judge/gateway/internal/logger"
	"github.com/nathan-foo/online-judge/gateway/internal/proxy"
	"github.com/nathan-foo/online-judge/gateway/internal/ratelimit"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog/log"
)

func NewMux(allowedOrigins []string, routes []config.RouteConfig,
	authn *auth.Middleware, rl *ratelimit.RateLimiter, hc *health.Checker) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(logger.RequestLogger)
	r.Use(middleware.Recoverer)

	r.Group(func(r chi.Router) {
		r.Get("/healthz", hc.Healthz)
		r.Get("/readyz", hc.Readyz)
	})

	r.Group(func(r chi.Router) {
		r.Use(securityHeaders)
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   allowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowedHeaders:   []string{"Authorization", "Content-Type"},
			AllowCredentials: true,
			MaxAge:           3600,
		}))
		r.Use(middleware.Timeout(10 * time.Second))
		r.Use(rl.Global())

		for _, route := range routes {
			handler, err := proxy.NewProxy(route.ServiceUrl, route.Prefix)
			if err != nil {
				log.Fatal().Err(err).Str("route", route.Prefix).Msg("failed to create proxy")
			}
			r.Route(route.Prefix, func(r chi.Router) {
				if route.RequireAuth {
					r.Use(authn.WithAuth)
				}
				if route.RateLimit > 0 {
					r.Use(rl.Route(route.Prefix, route.RateLimit))
				}
				if route.MaxUploadSize > 0 {
					r.Use(uploadHandler(route.MaxUploadSize))
				}
				r.Mount("/", handler)
			})
		}
	})

	return r
}

func uploadHandler(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("X-User-ID")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		// w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}
