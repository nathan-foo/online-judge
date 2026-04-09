package ratelimit

import (
	"net/http"
	"time"

	"github.com/nathan-foo/online-judge/gateway/internal/config"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	rdb    *redis.Client
	global *SlidingWindowLimiter
}

func NewRateLimiter(rdb *redis.Client, redisConfig config.RedisConfig) *RateLimiter {
	return &RateLimiter{
		rdb:    rdb,
		global: NewSlidingWindowLimiter(rdb, redisConfig.RateLimitGlobal, time.Minute),
	}
}

func (rl *RateLimiter) Global() func(http.Handler) http.Handler {
	return newMiddleware(rl.global, ipKey)
}

func (rl *RateLimiter) Route(prefix string, limit int) func(http.Handler) http.Handler {
	limiter := NewSlidingWindowLimiter(rl.rdb, limit, time.Minute)
	return newMiddleware(limiter, routeKey(prefix))
}
