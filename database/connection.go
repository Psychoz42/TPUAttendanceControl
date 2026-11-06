package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

func Connection(connStr string) error {
	var err error

	pool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		pool = nil
		return fmt.Errorf("ping database: %w", err)
	}

	fmt.Println("Database connected successfully")
	return nil
}

func GetPool() *pgxpool.Pool {
	return pool
}
