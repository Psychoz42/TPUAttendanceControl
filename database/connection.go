package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"fmt"
	"log"
)

var pool *pgxpool.Pool

func Connection() {
	connStr := "postgresql://postgres:2367@localhost:5431/postgres"
	var err error

	pool, err = pgxpool.New(context.Background(), connStr)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database connected successfully")

}

func CheckConnection() {
	err := pool.Ping(context.Background())
	if err != nil {
		fmt.Println("db is not connected")
	} else {
		fmt.Println("db is connected")
	}
}

func GetPool() *pgxpool.Pool {
	return pool
}
