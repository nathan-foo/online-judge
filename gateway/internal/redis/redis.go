package redis

import (
	"context"

	"github.com/nathan-foo/online-judge/gateway/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewClient(redisConfig config.RedisConfig) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisConfig.RedisUrl)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opt)

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		rdb.Close()
		return nil, err
	}

	return rdb, nil
}
