package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config represents strongly-typed platform configuration.
type Config struct {
	// Application Environment & Server
	AppEnv  string `env:"NODE_ENV" envDefault:"development"`
	Host    string `env:"HOST" envDefault:"0.0.0.0"`
	Port    int    `env:"PORT" envDefault:"3000"`
	APIBase string `env:"API_PREFIX" envDefault:"/api"`

	// Database Connection (Critical Secret)
	DatabaseURL string `env:"DATABASE_URL" envDefault:""`

	// Authentication & Security (Critical Secrets)
	JWTSecret           string        `env:"JWT_SECRET" envDefault:""`
	JWTRefreshSecret    string        `env:"JWT_REFRESH_SECRET" envDefault:""`
	JWTExpiresIn        time.Duration `env:"JWT_EXPIRES_IN" envDefault:"15m"`
	JWTRefreshExpiresIn time.Duration `env:"JWT_REFRESH_EXPIRES_IN" envDefault:"7d"`
	EncryptionKey       string        `env:"ENCRYPTION_KEY" envDefault:"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"`

	// Single Sign-On (SSO) / OpenID Connect (OIDC)
	SSOEnabled      bool   `env:"SSO_ENABLED" envDefault:"false"`
	SSOProviderName string `env:"SSO_PROVIDER_NAME" envDefault:"Google"`
	SSOIssuerURL    string `env:"SSO_ISSUER_URL" envDefault:"https://accounts.google.com"`
	SSOClientID     string `env:"SSO_CLIENT_ID" envDefault:""`
	SSOClientSecret string `env:"SSO_CLIENT_SECRET" envDefault:""`
	SSORedirectURI  string `env:"SSO_REDIRECT_URI" envDefault:"http://localhost:3000/api/auth/oidc/callback"`

	// Redis Configuration
	RedisHost     string `env:"REDIS_HOST" envDefault:"localhost"`
	RedisPort     int    `env:"REDIS_PORT" envDefault:"6379"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`

	// SMTP Email Configuration
	SMTPHost    string `env:"SMTP_HOST" envDefault:"smtp.example.com"`
	SMTPPort    int    `env:"SMTP_PORT" envDefault:"587"`
	SMTPUser    string `env:"SMTP_USER" envDefault:"noreply@orbito.io"`
	SMTPPass    string `env:"SMTP_PASS" envDefault:""`
	SMTPFrom    string `env:"SMTP_FROM" envDefault:"noreply@orbito.io"`
	EmailDomain string `env:"EMAIL_DOMAIN" envDefault:"orbito.io"`

	// Client URLs & CORS
	FrontendURL string   `env:"FRONTEND_URL" envDefault:"http://localhost:3001"`
	CORSOrigin  string   `env:"CORS_ORIGIN" envDefault:"http://localhost:3001"`
	CORSOrigins []string `env:"CORS_ORIGINS" envDefault:"http://localhost:3000,http://localhost:3001,http://127.0.0.1:3000,http://127.0.0.1:3001"`

	// Frontend Public URL References
	NextPublicAPIBaseURL string `env:"NEXT_PUBLIC_API_BASE_URL" envDefault:"http://localhost:3000/api"`
	NextPublicSocketURL  string `env:"NEXT_PUBLIC_SOCKET_URL" envDefault:"http://localhost:3000"`
	NextPublicAppURL     string `env:"NEXT_PUBLIC_APP_URL" envDefault:"http://localhost:3001"`

	// File Upload Settings
	UploadDest              string   `env:"UPLOAD_DEST" envDefault:"./uploads"`
	MaxFileSize             int64    `env:"MAX_FILE_SIZE" envDefault:"10485760"`
	MaxEditorImageSize      int64    `env:"MAX_EDITOR_IMAGE_SIZE" envDefault:"5242880"`
	AllowedImageTypes       []string `env:"ALLOWED_IMAGE_TYPES" envDefault:"image/jpeg,image/png,image/gif,image/webp"`
	EditorImageUploadFolder string   `env:"EDITOR_IMAGE_UPLOAD_FOLDER" envDefault:"editor-images"`
	MaxAvatarSize           int64    `env:"MAX_AVATAR_SIZE" envDefault:"52428800"`

	// Queue Worker Concurrency
	MaxConcurrentJobs int    `env:"MAX_CONCURRENT_JOBS" envDefault:"5"`
	JobRetryAttempts  int    `env:"JOB_RETRY_ATTEMPTS" envDefault:"3"`
	BullBoardEmail    string `env:"BULL_BOARD_EMAIL" envDefault:"admin@orbito.io"`
	BullBoardPassword string `env:"BULL_BOARD_PASSWORD" envDefault:"admin_secret_password"`

	// SSRF Security & Integrations
	JiraAllowedHosts        []string `env:"JIRA_ALLOWED_HOSTS" envDefault:"*.atlassian.net"`
	AIAllowedHosts          []string `env:"AI_ALLOWED_HOSTS" envDefault:"api.openai.com,*.anthropic.com,openrouter.ai"`
	AIAllowPrivateEndpoints bool     `env:"AI_ALLOW_PRIVATE_ENDPOINTS" envDefault:"false"`

	// AWS S3 Cloud Storage (Optional)
	AWSAccessKeyID     string `env:"AWS_ACCESS_KEY_ID" envDefault:""`
	AWSSecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY" envDefault:""`
	AWSRegion          string `env:"AWS_REGION" envDefault:"us-east-1"`
	AWSBucketName      string `env:"AWS_BUCKET_NAME" envDefault:"orbito-assets"`
}

// IsProduction returns true if NODE_ENV is production.
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.AppEnv) == "production"
}

// ParseCustomDuration parses standard Go durations plus days suffix ('d' or 'D').
func ParseCustomDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty duration string")
	}

	if strings.HasSuffix(s, "d") || strings.HasSuffix(s, "D") {
		numStr := s[:len(s)-1]
		days, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, fmt.Errorf("invalid days format %q: %w", s, err)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	return time.ParseDuration(s)
}

// Load loads and validates configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{}

	opts := env.Options{
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeOf(time.Duration(0)): func(v string) (any, error) {
				return ParseCustomDuration(v)
			},
		},
	}

	if err := env.ParseWithOptions(cfg, opts); err != nil {
		return nil, fmt.Errorf("failed to parse environment configuration: %w", err)
	}

	// Validate critical configuration fields
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return nil, errors.New("missing critical configuration: DATABASE_URL must not be empty")
	}

	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return nil, errors.New("missing critical configuration: JWT_SECRET must not be empty")
	}

	// Default JWTRefreshSecret to JWTSecret if not provided
	if strings.TrimSpace(cfg.JWTRefreshSecret) == "" {
		cfg.JWTRefreshSecret = cfg.JWTSecret
	}

	return cfg, nil
}

// MustLoad loads configuration or panics if critical variables are missing or invalid.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("CRITICAL CONFIG ERROR: %v", err))
	}
	return cfg
}
