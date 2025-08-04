package database

import (
	"database/sql"
	"fmt"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

// NewConnection creates a new PostgreSQL database connection
func NewConnection(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

// NewRedisConnection creates a new Redis connection
func NewRedisConnection(redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test the connection
	_, err = client.Ping(client.Context()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return client, nil
}