package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Gurveer1510/urlshortner/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type DB struct {
	Pool *pgxpool.Pool
}

func DSN(conf *config.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", conf.DBUser, conf.DBPass, conf.DBHost, conf.DBName, conf.SSL)
}

func NewPool(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	if err := runMigrations(dsn); err != nil {
		return nil, err
	}

	return &DB{Pool: pool}, nil
}

func runMigrations(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("sql.Open: %w", err)
	}

	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate.NewWithDatabaseInstance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate Up: %w", err)
	}

	log.Println("migrations applied")
	return nil

}
