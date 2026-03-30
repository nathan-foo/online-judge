package router

import (
	"net/http"
	"time"

	"github.com/nathan-foo/online-judge/gateway/internal/auth"
	"github.com/nathan-foo/online-judge/gateway/internal/proxy"
	"github.com/nathan-foo/online-judge/gateway/internal/ratelimit"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewMux(allowedOrigins []string, authn *auth.Middleware, p *proxy.Proxy, rl *ratelimit.RateLimiter) *chi.Mux {
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

	r.Route("/test", func(r chi.Router) {
		r.Use(authn.WithAuth)
		r.Use(rl.Route())
		r.Use(UploadHandler(1 << 20))
		r.Mount("/", http.StripPrefix("/test", p.TestHandler()))
	})

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
