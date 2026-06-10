package database

import (
	"context"

	"github.com/nathan-foo/online-judge/attempt-service/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(databaseConfig config.DatabaseConfig) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), databaseConfig.DatabaseUrl)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
