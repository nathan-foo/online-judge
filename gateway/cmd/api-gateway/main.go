package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nathan-foo/online-judge/gateway/internal/auth"
	"github.com/nathan-foo/online-judge/gateway/internal/config"
	"github.com/nathan-foo/online-judge/gateway/internal/ratelimit"
	"github.com/nathan-foo/online-judge/gateway/internal/redis"
	"github.com/nathan-foo/online-judge/gateway/internal/router"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	rdb, err := redis.NewClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Error connecting to redis: %v", err)
	}
	defer rdb.Close()

	rl := ratelimit.NewRateLimiter(rdb, cfg.Redis)

	authn := auth.NewMiddleware(cfg.Auth)

	mux := router.NewMux(cfg.AllowedOrigins, cfg.Routes, authn, rl)
	server := http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 25 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Server running on port :%d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error running http server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Error shutting down server: %v", err)
	}

	log.Println("Server stopped")
}
