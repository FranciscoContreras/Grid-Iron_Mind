// Package db provides database connection pooling and management for PostgreSQL.
//
// This package handles all database connectivity, connection pooling configuration,
// health checks, and pool metrics monitoring. It uses pgx/v5 for efficient PostgreSQL
// connection pooling with automatic reconnection and health validation.
//
// Key Features:
//   - Connection pooling with configurable min/max connections
//   - Automatic connection health checks and validation
//   - Pool metrics and monitoring (acquire time, utilization, etc.)
//   - Graceful degradation on connection failures
//   - Request timeout and lifecycle management
//
// Example usage:
//
//	cfg := db.Config{
//	    DatabaseURL: "postgres://user:pass@localhost/db",
//	    MaxConns:    25,
//	    MinConns:    5,
//	}
//	if err := db.Connect(ctx, cfg); err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
//
//	pool := db.GetPool()
//	rows, err := pool.Query(ctx, "SELECT * FROM users")
package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

// Config holds database connection pool configuration.
//
// DatabaseURL should be in the format:
// postgres://username:password@host:port/database?sslmode=require
//
// MaxConns and MinConns control the connection pool size.
// Recommended: MaxConns = (CPU cores * 2) + effective_spindle_count
type Config struct {
	DatabaseURL string // PostgreSQL connection string
	MaxConns    int32  // Maximum number of connections in pool (default: 25)
	MinConns    int32  // Minimum number of idle connections (default: 5)
}

// Connect establishes a connection pool to PostgreSQL with the given configuration.
//
// This function:
//   - Parses the database URL and creates a pool configuration
//   - Sets connection timeouts and lifecycle hooks
//   - Validates connections before acquisition (BeforeAcquire hook)
//   - Configures health check intervals
//   - Verifies connectivity with an initial ping
//
// Returns an error if the database URL is invalid or connection fails.
// The pool is stored globally and can be accessed via GetPool().
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

	// Set connection timeout and acquire timeout
	config.ConnConfig.ConnectTimeout = 10 * time.Second
	config.MaxConnIdleTime = 5 * time.Minute

	// Configure pool behavior
	config.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		// Validate connection is still good before using
		return conn.Ping(ctx) == nil
	}

	config.AfterRelease = func(conn *pgx.Conn) bool {
		// Return true to keep connection in pool
		// Return false to close connection
		return true
	}

	// Create connection pool
	pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	log.Printf("Successfully connected to PostgreSQL (MaxConns: %d, MinConns: %d)", cfg.MaxConns, cfg.MinConns)
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

// PoolMetrics returns detailed pool metrics as a map
func PoolMetrics() map[string]interface{} {
	if pool == nil {
		return map[string]interface{}{
			"error": "pool not initialized",
		}
	}

	stat := pool.Stat()
	return map[string]interface{}{
		"acquired_conns":       stat.AcquiredConns(),
		"idle_conns":           stat.IdleConns(),
		"max_conns":            stat.MaxConns(),
		"total_conns":          stat.TotalConns(),
		"new_conns_count":      stat.NewConnsCount(),
		"acquire_count":        stat.AcquireCount(),
		"acquire_duration_ms":  stat.AcquireDuration().Milliseconds(),
		"empty_acquire_count":  stat.EmptyAcquireCount(),
		"canceled_acquire_count": stat.CanceledAcquireCount(),
	}
}

// LogPoolStats logs current pool statistics
func LogPoolStats() {
	metrics := PoolMetrics()
	log.Printf("[DB-POOL] Acquired: %v, Idle: %v, Max: %v, Total: %v, Acquire Duration: %vms",
		metrics["acquired_conns"],
		metrics["idle_conns"],
		metrics["max_conns"],
		metrics["total_conns"],
		metrics["acquire_duration_ms"],
	)
}

// IsHealthy checks if pool is in healthy state
func IsHealthy() bool {
	if pool == nil {
		return false
	}

	stat := pool.Stat()

	// Unhealthy if all connections are acquired (pool exhaustion)
	if stat.AcquiredConns() >= stat.MaxConns() {
		log.Printf("[DB-POOL] WARNING: Pool exhaustion - %d/%d connections acquired",
			stat.AcquiredConns(), stat.MaxConns())
		return false
	}

	// Unhealthy if acquire duration is too high (>100ms)
	if stat.AcquireDuration().Milliseconds() > 100 {
		log.Printf("[DB-POOL] WARNING: High acquire duration - %dms",
			stat.AcquireDuration().Milliseconds())
		return false
	}

	// Unhealthy if many canceled acquisitions (connection timeout)
	if stat.CanceledAcquireCount() > 10 {
		log.Printf("[DB-POOL] WARNING: High canceled acquire count - %d",
			stat.CanceledAcquireCount())
		return false
	}

	return true
}