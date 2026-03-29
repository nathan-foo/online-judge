package ratelimit

import (
	"net/http"
	"time"

	"github.com/nathan-foo/online-judge/gateway/internal/config"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	global *SlidingWindowLimiter
	route  *SlidingWindowLimiter
}

func NewRateLimiter(rdb *redis.Client, redisConfig config.RedisConfig) *RateLimiter {
	return &RateLimiter{
		global: NewSlidingWindowLimiter(rdb, redisConfig.RateLimitGlobal, time.Minute),
		route:  NewSlidingWindowLimiter(rdb, redisConfig.RateLimitRoute, time.Minute),
	}
}

func (rl *RateLimiter) Global() func(http.Handler) http.Handler {
	return newMiddleware(rl.global, ipKey)
}

func (rl *RateLimiter) Route() func(http.Handler) http.Handler {
	return newMiddleware(rl.route, userKey)
}
