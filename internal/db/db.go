package db

import (
	"context"
	"log/slog"
	"os"
	"regexp"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	redactedDSN := regexp.MustCompile(`://[^:]+:[^@]+@`).ReplaceAllString(dsn, "://***:***@")
	slog.Info("Connecting to DB", slog.String("dsn", redactedDSN))

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		slog.Error("Unable to connect to DB", slog.Any("error", err))
		os.Exit(1)
	}

	DB = pool
	slog.Info("Connected to DB")
}
