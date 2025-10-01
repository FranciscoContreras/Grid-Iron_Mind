package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var client *redis.Client

// Config holds Redis configuration
type Config struct {
	RedisURL string
}

// Connect establishes a connection to Redis
func Connect(cfg Config) error {
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Configure TLS for Heroku Redis (skip hostname verification)
	// Heroku Redis uses SSL but the certificate hostname may not match the actual endpoint
	if opt.TLSConfig != nil {
		opt.TLSConfig.InsecureSkipVerify = true
	} else if len(cfg.RedisURL) > 8 && cfg.RedisURL[:8] == "rediss://" {
		// If using rediss:// protocol but TLSConfig not set, create one
		opt.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	client = redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Successfully connected to Redis")
	return nil
}

// Close closes the Redis connection
func Close() error {
	if client != nil {
		return client.Close()
	}
	return nil
}

// GetClient returns the Redis client
func GetClient() *redis.Client {
	return client
}

// Get retrieves a value from cache
func Get(ctx context.Context, key string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("Redis client not initialized")
	}

	val, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Key doesn't exist
	}
	if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}

	return val, nil
}

// Set stores a value in cache with TTL
func Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	err := client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// Delete removes a key from cache
func Delete(ctx context.Context, key string) error {
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	err := client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	return nil
}

// DeletePattern deletes all keys matching a pattern
func DeletePattern(ctx context.Context, pattern string) error {
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	// Scan for matching keys
	iter := client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	// Delete matching keys
	if len(keys) > 0 {
		if err := client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
		log.Printf("Deleted %d cache keys matching pattern: %s", len(keys), pattern)
	}

	return nil
}

// Exists checks if a key exists in cache
func Exists(ctx context.Context, key string) (bool, error) {
	if client == nil {
		return false, fmt.Errorf("Redis client not initialized")
	}

	count, err := client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	return count > 0, nil
}

// GetWithRefresh retrieves a value and extends its TTL
func GetWithRefresh(ctx context.Context, key string, ttl time.Duration) (string, error) {
	val, err := Get(ctx, key)
	if err != nil {
		return "", err
	}

	if val != "" {
		// Extend TTL
		if err := client.Expire(ctx, key, ttl).Err(); err != nil {
			log.Printf("Warning: failed to refresh TTL for key %s: %v", key, err)
		}
	}

	return val, nil
}

// HealthCheck verifies Redis connectivity
func HealthCheck(ctx context.Context) error {
	if client == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}