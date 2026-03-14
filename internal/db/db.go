package db

import (
	"context"
	"fmt"

	"github.com/Gurveer1510/urlshortner/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func DSN(conf *config.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", conf.DBUser,conf.DBPass, conf.DBHost, conf.DBName, conf.SSL)
}

func NewPool(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &DB{Pool: pool}, nil

}
