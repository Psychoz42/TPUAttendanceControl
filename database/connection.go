package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"fmt"
)

var pool *pgxpool.Pool

func Connection() error {
	connStr := "postgresql://postgres:2367@localhost:5431/postgres"
	var err error

	pool, err = pgxpool.New(context.Background(), connStr)

	if err != nil {
		return err
	}

	fmt.Println("Database connected successfully")
	return nil
}

func CheckConnection() error {
	err := pool.Ping(context.Background())
	if err != nil {
		fmt.Println("db is not connected")
		return err
	} else {
		fmt.Println("db is connected")
	}

	return err
}

func GetPool() *pgxpool.Pool {
	return pool
}
