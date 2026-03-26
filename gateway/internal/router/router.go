package router

import (
	"fmt"
	"log"
	"net/http"

	"github.com/nathan-foo/online-judge/gateway/internal/auth"
	"github.com/nathan-foo/online-judge/gateway/internal/handler"
)

func NewMux(authn *auth.Middleware) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)

		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		fmt.Fprintf(w, "Server: %s %s", r.Method, r.URL.Path)
	})

	mux.Handle("/hello", authn.RequireSession(http.HandlerFunc(handler.HandleHello)))
	return mux
}
