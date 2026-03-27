package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nathan-foo/online-judge/gateway/internal/auth"
	"github.com/nathan-foo/online-judge/gateway/internal/proxy"
)

func NewMux(authn *auth.Middleware, p *proxy.Proxy) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Second * 60))

	r.Route("/test", func(r chi.Router) {
		r.With(authn.WithAuth).Mount("/", http.StripPrefix("/test", p.TestHandler()))
	})

	return r
}
