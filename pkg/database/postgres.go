package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

// Config holds database configuration
type Config struct {
	DatabaseURL string
	MaxConns    int32
	MinConns    int32
}

// Connect establishes a connection pool to PostgreSQL
func Connect(ctx context.Context, cfg Config) error {
	config, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Set connection pool settings
	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = time.Minute

	// Create connection pool
	pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL")
	return nil
}

// GetPool returns the database connection pool
func GetPool() *pgxpool.Pool {
	return pool
}

// Close closes the database connection pool
func Close() {
	if pool != nil {
		pool.Close()
		log.Println("Database connection pool closed")
	}
}

// HealthCheck verifies database connectivity
func HealthCheck(ctx context.Context) error {
	if pool == nil {
		return fmt.Errorf("database connection pool not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

// Stats returns connection pool statistics
func Stats() *pgxpool.Stat {
	if pool == nil {
		return nil
	}
	stat := pool.Stat()
	return stat
}