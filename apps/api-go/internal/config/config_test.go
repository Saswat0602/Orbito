package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCustomDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasErr   bool
	}{
		{"15m", 15 * time.Minute, false},
		{"1h", 1 * time.Hour, false},
		{"7d", 7 * 24 * time.Hour, false},
		{"14D", 14 * 24 * time.Hour, false},
		{"30s", 30 * time.Second, false},
		{"", 0, true},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseCustomDuration(tt.input)
			if tt.hasErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestConfigLoadDefaults(t *testing.T) {
	// Set critical secrets
	t.Setenv("JWT_SECRET", "test-jwt-secret-key-12345")
	t.Setenv("DATABASE_URL", "postgresql://root:root@localhost:5433/orbito")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify defaults
	assert.Equal(t, "development", cfg.AppEnv)
	assert.False(t, cfg.IsProduction())
	assert.Equal(t, 3000, cfg.Port)
	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, "/api", cfg.APIBase)
	assert.Equal(t, "test-jwt-secret-key-12345", cfg.JWTSecret)
	assert.Equal(t, "test-jwt-secret-key-12345", cfg.JWTRefreshSecret) // defaulted from JWT_SECRET
	assert.Equal(t, 15*time.Minute, cfg.JWTExpiresIn)
	assert.Equal(t, 7*24*time.Hour, cfg.JWTRefreshExpiresIn)
	assert.Equal(t, "localhost", cfg.RedisHost)
	assert.Equal(t, 6379, cfg.RedisPort)
	assert.Equal(t, "http://localhost:3001", cfg.FrontendURL)
	assert.Equal(t, "http://localhost:3001", cfg.CORSOrigin)
	assert.Equal(t, int64(10485760), cfg.MaxFileSize)
	assert.Equal(t, 5, cfg.MaxConcurrentJobs)
}

func TestConfigMissingJWTSecretFails(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("DATABASE_URL", "postgresql://root:root@localhost:5433/orbito")

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "JWT_SECRET must not be empty")

	assert.Panics(t, func() {
		MustLoad()
	})
}

func TestConfigMissingDatabaseURLFails(t *testing.T) {
	t.Setenv("JWT_SECRET", "super-secret-jwt")
	t.Setenv("DATABASE_URL", "")

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "DATABASE_URL must not be empty")

	assert.Panics(t, func() {
		MustLoad()
	})
}

func TestConfigOverrides(t *testing.T) {
	t.Setenv("JWT_SECRET", "custom-jwt-secret")
	t.Setenv("JWT_REFRESH_SECRET", "custom-refresh-secret")
	t.Setenv("DATABASE_URL", "postgresql://produser:prodpass@db.orbito.internal:5432/orbitodb")
	t.Setenv("NODE_ENV", "production")
	t.Setenv("PORT", "8080")
	t.Setenv("HOST", "127.0.0.1")
	t.Setenv("JWT_EXPIRES_IN", "30m")
	t.Setenv("JWT_REFRESH_EXPIRES_IN", "30d")
	t.Setenv("REDIS_HOST", "redis.orbito.internal")
	t.Setenv("REDIS_PORT", "6380")
	t.Setenv("FRONTEND_URL", "https://orbito.io")
	t.Setenv("CORS_ORIGIN", "https://orbito.io")
	t.Setenv("SSO_ENABLED", "true")
	t.Setenv("MAX_CONCURRENT_JOBS", "10")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "production", cfg.AppEnv)
	assert.True(t, cfg.IsProduction())
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, "custom-jwt-secret", cfg.JWTSecret)
	assert.Equal(t, "custom-refresh-secret", cfg.JWTRefreshSecret)
	assert.Equal(t, 30*time.Minute, cfg.JWTExpiresIn)
	assert.Equal(t, 30*24*time.Hour, cfg.JWTRefreshExpiresIn)
	assert.Equal(t, "redis.orbito.internal", cfg.RedisHost)
	assert.Equal(t, 6380, cfg.RedisPort)
	assert.Equal(t, "https://orbito.io", cfg.FrontendURL)
	assert.Equal(t, "https://orbito.io", cfg.CORSOrigin)
	assert.True(t, cfg.SSOEnabled)
	assert.Equal(t, 10, cfg.MaxConcurrentJobs)
}
