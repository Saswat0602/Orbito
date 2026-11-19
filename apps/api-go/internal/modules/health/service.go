package health

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// ServiceHealth describes the operational state of a dependency service.
type ServiceHealth struct {
	Status       string `json:"status"` // "up" or "down"
	ResponseTime *int64 `json:"responseTime,omitempty"`
	Error        string `json:"error,omitempty"`
}

// ServicesHealth groups dependencies checked by the health subsystem.
type ServicesHealth struct {
	Database ServiceHealth `json:"database"`
	Redis    ServiceHealth `json:"redis"`
}

// Status matches the NestJS HealthStatus interface.
type Status struct {
	Status    string         `json:"status"` // "healthy" or "unhealthy"
	Timestamp string         `json:"timestamp"`
	Uptime    int64          `json:"uptime"`
	Services  ServicesHealth `json:"services"`
}

// LivenessStatus matches the NestJS liveness response.
type LivenessStatus struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// Service coordinates health checks across dependencies.
type Service struct {
	startTime time.Time
	db        *database.DB
	cfg       *config.Config
}

// NewService creates a new HealthService.
func NewService(db *database.DB, cfg *config.Config) *Service {
	return &Service{
		startTime: time.Now(),
		db:        db,
		cfg:       cfg,
	}
}

// CheckDatabase executes a ping against PostgreSQL and measures latency in milliseconds.
func (s *Service) CheckDatabase(ctx context.Context) ServiceHealth {
	start := time.Now()
	if s.db == nil {
		duration := time.Since(start).Milliseconds()
		return ServiceHealth{
			Status:       "down",
			ResponseTime: &duration,
			Error:        "database pool not initialized",
		}
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if pingErr := s.db.Ping(pingCtx); pingErr != nil {
		duration := time.Since(start).Milliseconds()
		return ServiceHealth{
			Status:       "down",
			ResponseTime: &duration,
			Error:        pingErr.Error(),
		}
	}

	duration := time.Since(start).Milliseconds()
	return ServiceHealth{
		Status:       "up",
		ResponseTime: &duration,
	}
}

// CheckRedis executes a lightweight RESP ping against Redis and measures latency in milliseconds.
func (s *Service) CheckRedis(ctx context.Context) ServiceHealth {
	start := time.Now()

	redisHost := "localhost"
	redisPort := 6379
	redisPassword := ""

	if s.cfg != nil {
		if s.cfg.RedisHost != "" {
			redisHost = s.cfg.RedisHost
		}
		if s.cfg.RedisPort != 0 {
			redisPort = s.cfg.RedisPort
		}
		redisPassword = s.cfg.RedisPassword
	}

	addr := fmt.Sprintf("%s:%d", redisHost, redisPort)
	dialer := net.Dialer{Timeout: 2 * time.Second}
	conn, dialErr := dialer.DialContext(ctx, "tcp", addr)
	if dialErr != nil {
		duration := time.Since(start).Milliseconds()
		return ServiceHealth{
			Status:       "down",
			ResponseTime: &duration,
			Error:        dialErr.Error(),
		}
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))

	if redisPassword != "" {
		if _, authErr := fmt.Fprintf(conn, "*2\r\n$4\r\nAUTH\r\n$%d\r\n%s\r\n", len(redisPassword), redisPassword); authErr != nil {
			duration := time.Since(start).Milliseconds()
			return ServiceHealth{
				Status:       "down",
				ResponseTime: &duration,
				Error:        authErr.Error(),
			}
		}
		reader := bufio.NewReader(conn)
		line, readErr := reader.ReadString('\n')
		if readErr != nil || (!strings.HasPrefix(line, "+OK") && !strings.HasPrefix(line, "+ok")) {
			duration := time.Since(start).Milliseconds()
			return ServiceHealth{
				Status:       "down",
				ResponseTime: &duration,
				Error:        fmt.Sprintf("redis auth failed: %s", strings.TrimSpace(line)),
			}
		}
	}

	if _, pingErr := fmt.Fprintf(conn, "*1\r\n$4\r\nPING\r\n"); pingErr != nil {
		duration := time.Since(start).Milliseconds()
		return ServiceHealth{
			Status:       "down",
			ResponseTime: &duration,
			Error:        pingErr.Error(),
		}
	}

	reader := bufio.NewReader(conn)
	line, readErr := reader.ReadString('\n')
	if readErr != nil || (!strings.HasPrefix(line, "+PONG") && !strings.HasPrefix(line, "+pong")) {
		duration := time.Since(start).Milliseconds()
		return ServiceHealth{
			Status:       "down",
			ResponseTime: &duration,
			Error:        fmt.Sprintf("redis ping failed: %s", strings.TrimSpace(line)),
		}
	}

	duration := time.Since(start).Milliseconds()
	return ServiceHealth{
		Status:       "up",
		ResponseTime: &duration,
	}
}

// CheckHealth runs concurrent health checks across database and redis and aggregates status.
func (s *Service) CheckHealth(ctx context.Context) Status {
	var wg sync.WaitGroup
	var dbHealth ServiceHealth
	var redisHealth ServiceHealth

	wg.Add(2)
	go func() {
		defer wg.Done()
		dbHealth = s.CheckDatabase(ctx)
	}()

	go func() {
		defer wg.Done()
		redisHealth = s.CheckRedis(ctx)
	}()

	wg.Wait()

	skipRedis := os.Getenv("SKIP_REDIS_CHECK") == "true"
	isHealthy := dbHealth.Status == "up" && (redisHealth.Status == "up" || skipRedis)

	statusStr := "unhealthy"
	if isHealthy {
		statusStr = "healthy"
	}

	uptimeSeconds := int64(time.Since(s.startTime).Seconds())

	return Status{
		Status:    statusStr,
		Timestamp: httpx.FormatTimestamp(),
		Uptime:    uptimeSeconds,
		Services: ServicesHealth{
			Database: dbHealth,
			Redis:    redisHealth,
		},
	}
}
