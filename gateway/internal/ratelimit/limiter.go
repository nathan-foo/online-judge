package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SlidingWindowLimiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
}

type Result struct {
	Allowed   bool
	Remaining int
}

func NewSlidingWindowLimiter(rdb *redis.Client, limit int, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		rdb:    rdb,
		limit:  limit,
		window: window,
	}
}

var script = redis.NewScript(`
local current_key = KEYS[1]
local previous_key = KEYS[2]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local current_start = now - (now % window)
local weight = (now - current_start) / window

local previous_count = tonumber(redis.call('GET', previous_key) or '0')
local current_count = tonumber(redis.call('GET', current_key) or '0')
local weighted_count = previous_count * (1 - weight) + current_count

if weighted_count >= limit then
	return -1
end

redis.call('INCR', current_key)
redis.call('EXPIRE', current_key, window * 2)
return limit - math.ceil(weighted_count) - 1
`)

func (l *SlidingWindowLimiter) Allow(ctx context.Context, key string) (*Result, error) {
	now := time.Now()
	windowSeconds := int(l.window.Seconds())
	windowStart := now.Truncate(l.window).Unix()

	currentKey := fmt.Sprintf("%s:%d", key, windowStart)
	previousKey := fmt.Sprintf("%s:%d", key, windowStart-int64(windowSeconds))

	result, err := script.Run(ctx, l.rdb,
		[]string{currentKey, previousKey},
		l.limit,
		windowSeconds,
		now.Unix(),
	).Int()

	if err != nil {
		return nil, err
	}

	if result == -1 {
		return &Result{Allowed: false, Remaining: 0}, nil
	}

	return &Result{Allowed: true, Remaining: result}, nil
}
