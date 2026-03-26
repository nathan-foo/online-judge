package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nathan-foo/online-judge/gateway/internal/auth"
	"github.com/nathan-foo/online-judge/gateway/internal/config"
	"github.com/nathan-foo/online-judge/gateway/internal/router"
)

const serverPort = 8080

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config")
	}

	authn, err := auth.NewMiddleware(cfg)
	if err != nil {
		log.Fatalf("Error authenticating")
	}

	mux := router.NewMux(authn)
	server := http.Server{
		Addr:         fmt.Sprintf(":%d", serverPort),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Server running on port :%d", serverPort)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Error running http server: %v", err)
	}
}
