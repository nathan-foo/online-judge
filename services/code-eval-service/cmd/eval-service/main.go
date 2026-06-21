package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/broker"
	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/config"
	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/executor"
	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/health"
	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/judge"
	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/logger"
	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/pool"
	"github.com/nathan-foo/online-judge/services/code-eval-service/internal/router"
)

const evalTimeout = 90 * time.Second

func main() {
	logger.Init()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	images := make(map[judge.Language]string, len(cfg.Languages))
	sizes := make(map[judge.Language]int, len(cfg.Languages))
	totalSlots := 0
	for _, l := range cfg.Languages {
		lang := judge.Language(l)
		images[lang] = fmt.Sprintf("%s-%s:%s", cfg.ExecAgentImage, l, cfg.ExecAgentTag)
		size := cfg.PoolSize
		if s, ok := cfg.PoolSizes[l]; ok {
			size = s
		}
		sizes[lang] = size
		totalSlots += size
	}

	prefetch := totalSlots
	if prefetch < 1 {
		prefetch = len(images)
	}

	b := &broker.Broker{}
	if err := b.Connect(cfg.RabbitMQURL, prefetch); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to rabbitmq")
	}

	k8s, err := pool.NewClient()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create kubernetes client")
	}
	mgr := pool.NewManager(pool.ManagerOptions{
		Client:       k8s,
		Namespace:    cfg.Namespace,
		Images:       images,
		Sizes:        sizes,
		RuntimeClass: cfg.RuntimeClassName,
	})
	startCtx, startCancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := mgr.Start(startCtx); err != nil {
		startCancel()
		log.Fatal().Err(err).Msg("failed to start exec pool")
	}
	startCancel()

	exe := executor.NewClient()

	handler := func(body []byte) error {
		var req judge.EvalRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Error().Err(err).Msg("failed to decode eval request")
			return fmt.Errorf("decode eval request: %w", broker.ErrPoison)
		}

		result := evaluate(mgr, exe, req)

		payload, err := json.Marshal(result)
		if err != nil {
			log.Error().Err(err).Str("attempt_id", req.AttemptID).Msg("failed to encode eval result")
			return fmt.Errorf("encode eval result: %w", broker.ErrPoison)
		}
		if err := b.PublishResult(context.Background(), payload); err != nil {
			log.Error().Err(err).Str("attempt_id", req.AttemptID).Msg("failed to publish eval result")
			return err
		}
		return nil
	}

	if err := b.ConsumeRequests(handler); err != nil {
		log.Fatal().Err(err).Msg("failed to start consumer")
	}

	hc := health.NewChecker(
		health.Check{Name: "rabbitmq", Fn: func(ctx context.Context) error {
			return b.Healthy()
		}},
		health.Check{Name: "pool", Fn: func(ctx context.Context) error {
			return mgr.Ready()
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

	mgr.Shutdown(shutdownCtx)

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("server shutdown error")
	}

	log.Info().Msg("server stopped")
}

func evaluate(mgr *pool.Manager, exe *executor.Client, req judge.EvalRequest) judge.EvalResult {
	ctx, cancel := context.WithTimeout(context.Background(), evalTimeout)
	defer cancel()

	pod, err := mgr.Lease(ctx, req.Language)
	if err != nil {
		log.Error().Err(err).
			Str("attempt_id", req.AttemptID).
			Str("language", string(req.Language)).
			Msg("failed to lease exec pod")
		return judge.Aggregate(req, nil, "", err)
	}
	defer mgr.Release(req.Language, pod)

	resp, err := exe.Run(ctx, pod.URL, req)
	if err != nil {
		log.Error().Err(err).
			Str("attempt_id", req.AttemptID).
			Str("pod", pod.Name).
			Msg("exec failed")
		return judge.Aggregate(req, nil, "", err)
	}
	return judge.Aggregate(req, resp.Results, resp.CompileError, nil)
}
