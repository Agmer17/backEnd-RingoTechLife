package configs

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setUpDatabase(ctx context.Context, url string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 15
	config.MinConns = 3
	config.MaxConnIdleTime = 20 * time.Minute
	config.MaxConnLifetime = 10 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return pool, nil

}
