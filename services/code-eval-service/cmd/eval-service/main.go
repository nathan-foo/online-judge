package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/broker"
	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/config"
	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/health"
	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/logger"
	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/router"
)

func main() {
	logger.Init()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	b := &broker.Broker{}
	if err := b.Connect(cfg.RabbitMQURL); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to rabbitmq")
	}

	handler := func(body []byte) error {
		result := []byte{}

		// TODO

		return b.PublishResult(context.Background(), result)
	}
	if err := b.ConsumeRequests(handler); err != nil {
		log.Fatal().Err(err).Msg("failed to start consumer")
	}

	hc := health.NewChecker(
		health.Check{Name: "rabbitmq", Fn: func(ctx context.Context) error {
			return b.Healthy()
		}},
	)

	mux := router.NewMux(hc)
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
		log.Info().Str("addr", server.Addr).Msg("server started")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("http server error")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("shutting down server")

	if err := b.Close(); err != nil {
		log.Error().Err(err).Msg("broker shutdown error")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("server shutdown error")
	}

	log.Info().Msg("server stopped")
}
