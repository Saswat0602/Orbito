package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// DB manages the pgx connection pool and provides access to compiled sqlc queries.
type DB struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// Config holds pool sizing and connection lifecycle options.
type Config struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// DefaultConfig provides production-ready pool defaults matching the migration plan.
func DefaultConfig() Config {
	return Config{
		MaxConns:          50,
		MinConns:          10,
		MaxConnLifetime:   1 * time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
	}
}

// New initializes a new PostgreSQL connection pool and sqlc query handle.
func New(ctx context.Context, databaseURL string, cfg ...Config) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database connection URL: %w", err)
	}

	c := DefaultConfig()
	if len(cfg) > 0 {
		c = cfg[0]
	}

	poolConfig.MaxConns = c.MaxConns
	poolConfig.MinConns = c.MinConns
	poolConfig.MaxConnLifetime = c.MaxConnLifetime
	poolConfig.MaxConnIdleTime = c.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = c.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgxpool: %w", err)
	}

	// Verify connection with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{
		pool:    pool,
		queries: sqlc.New(pool),
	}, nil
}

// NewFromPool wraps an existing pgxpool.Pool (useful for integration testing and mocks).
func NewFromPool(pool *pgxpool.Pool) *DB {
	return &DB{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// Ping verifies the database connection is alive.
func (db *DB) Ping(ctx context.Context) error {
	if db.pool == nil {
		return fmt.Errorf("database pool is not initialized")
	}
	return db.pool.Ping(ctx)
}

// Close gracefully terminates the connection pool.
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// Pool returns the underlying pgx connection pool.
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// Queries returns the compiled sqlc queries.
func (db *DB) Queries() *sqlc.Queries {
	return db.queries
}
