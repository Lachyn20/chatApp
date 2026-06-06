package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

func NewPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Optimized connection pool settings
	db.SetMaxOpenConns(25)        // Increased for better concurrency
	db.SetMaxIdleConns(10)        // Keep more idle connections
	db.SetConnMaxLifetime(time.Hour)     // Recycle connections after 1 hour
	db.SetConnMaxIdleTime(time.Minute * 30) // Close idle connections after 30 minutes

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}