package router

import (
	"encoding/json"
	"net/http"

	"github.com/nathan-foo/online-judge/attempt-service/internal/health"
	"github.com/nathan-foo/online-judge/attempt-service/internal/logger"

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

	r.Group(func(r chi.Router) {
		r.Get("/", hello)
	})

	return r
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Hello from attempt-service"})
}
