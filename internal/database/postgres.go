package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saimonsiddique/blog-api/internal/config"
	"github.com/saimonsiddique/blog-api/internal/pkg/logger"
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

	// Log connection attempt with masked credentials
	logger.WithFields(map[string]interface{}{
		"host":     cfg.Host,
		"port":     cfg.Port,
		"database": cfg.Name,
		"user":     cfg.User,
		"sslmode":  cfg.SSLMode,
	}).Info("Connecting to PostgreSQL database")

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

	logger.WithFields(map[string]interface{}{
		"max_connections":     maxConnections,
		"min_connections":     minConnections,
		"max_conn_lifetime":   maxConnLifetime,
		"max_conn_idle_time":  maxConnIdleTime,
		"health_check_period": healthCheckPeriod,
	}).Debug("Database connection pool configuration")

	// Create connection pool with timeout
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Try to verify connection (non-fatal)
	if err := pool.Ping(ctx); err != nil {
		logger.WithError(err).Warn("Could not ping database - connection may not be working")
	} else {
		logger.Info("Database connection established successfully")
	}

	return pool, nil
}
