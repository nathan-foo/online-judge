package router

import (
	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/health"
	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewMux(hc *health.Checker) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(logger.RequestLogger)
	r.Use(middleware.Recoverer)

	r.Group(func(r chi.Router) {
		r.Get("/healthz", hc.Healthz)
		r.Get("/readyz", hc.Readyz)
	})

	return r
}
