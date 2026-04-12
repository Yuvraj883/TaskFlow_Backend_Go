package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	print("DSN: ", dsn)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("Unable to connect to DB:", err)
	}

	DB = pool
	log.Println("Connected to DB")
}
