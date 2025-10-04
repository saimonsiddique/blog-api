package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saimonsiddique/blog-api/internal/config"
)

const (
	maxConnections    = 25
	minConnections    = 5
	maxConnLifetime   = 5 * time.Minute
	maxConnIdleTime   = 1 * time.Minute
	healthCheckPeriod = 1 * time.Minute
	connectionTimeout = 5 * time.Second
)

func NewPostgresPool(cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode,
	)

	// Print a masked DSN for debugging
	// masked := fmt.Sprintf("postgres://%s:***@%s:%s/%s?sslmode=%s",
	// 	cfg.User, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode,
	// )

	// print the masked DSN
	fmt.Println("Connecting to database with DSN:", dsn)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Connection pool settings
	poolConfig.MaxConns = maxConnections
	poolConfig.MinConns = minConnections
	poolConfig.MaxConnLifetime = maxConnLifetime
	poolConfig.MaxConnIdleTime = maxConnIdleTime
	poolConfig.HealthCheckPeriod = healthCheckPeriod

	// Create connection pool with timeout
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Try to verify connection (non-fatal)
	if err := pool.Ping(ctx); err != nil {
		fmt.Printf("Warning: Could not ping database: %v\n", err)
		fmt.Println("Server will start but database connection may not be working")
	}

	return pool, nil
}
