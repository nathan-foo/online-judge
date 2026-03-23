package router

import (
	"net/http"

	"github.com/nathan-foo/online-judge/gateway/internal/handler"
)

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", handler.HandleHello)
	mux.HandleFunc("/goodbye", handler.HandleGoodbye)
	return mux
}
