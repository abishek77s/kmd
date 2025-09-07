package api

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitDB() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		log.Fatalf("Unable to parse DB URL: %v", err)
	}
	log.Println("Connecting to DB...")

	DB, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Unable to connect to DB: %v", err)
	}

	log.Println("Connected to DB")

}

// CREATE TABLE commands (
//     id TEXT PRIMARY KEY,
//     data JSONB NOT NULL
// );

// CREATE TABLE files (
//     id TEXT PRIMARY KEY,
//     data JSONB NOT NULL
// );
