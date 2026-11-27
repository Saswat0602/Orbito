package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hibiken/asynq"
	"github.com/orbito/api-go/internal/config"
	"github.com/redis/go-redis/v9"
)

const (
	// QueueAutomation is the background queue name for automation rules.
	QueueAutomation = "automation"
	// QueueBulkTaskImport is the background queue name for bulk task imports.
	QueueBulkTaskImport = "bulk-task-import"
	// QueueEmailSync is the background queue name for project inbox IMAP synchronization.
	QueueEmailSync = "email-sync"
	// TaskExecuteAutomationRule is the task type handled by the automation processor.
	TaskExecuteAutomationRule = "automation:execute-rule"
	// TaskBulkImportTasks is the task type handled by the bulk task import processor.
	TaskBulkImportTasks = "tasks:bulk-import"
	// TaskSyncProjectInbox is the task type handled by the project inbox email sync processor.
	TaskSyncProjectInbox = "inbox:sync-project"

	defaultInMemoryBuffer = 256
)

// HandlerFunc processes a background job payload.
type HandlerFunc func(ctx context.Context, payload []byte) error

type memoryJob struct {
	queueName string
	taskType  string
	payload   []byte
	attempt   int
}

// HealthStats holds aggregated queue execution counters.
type HealthStats struct {
	TotalJobs     int64   `json:"totalJobs"`
	CompletedJobs int64   `json:"completedJobs"`
	FailedJobs    int64   `json:"failedJobs"`
	ActiveJobs    int64   `json:"activeJobs"`
	SuccessRate   float64 `json:"successRate"`
}

// HealthResponse mirrors NestJS QueueHealthController output.
type HealthResponse struct {
	Backend          string      `json:"backend"`
	RequestedBackend string      `json:"requestedBackend"`
	FallbackOccurred bool        `json:"fallbackOccurred"`
	RedisAvailable   bool        `json:"redisAvailable"`
	Queues           []string    `json:"queues"`
	Stats            HealthStats `json:"stats"`
	Timestamp        string      `json:"timestamp"`
}

// Runtime manages asynq/Redis background queues with an automatic in-memory fallback pool.
type Runtime struct {
	log              *slog.Logger
	backend          string
	requestedBackend string
	fallbackOccurred bool
	redisAvailable   bool
	maxConcurrent    int
	maxRetries       int

	mu       sync.RWMutex
	queues   map[string]struct{}
	handlers map[string]HandlerFunc

	redisClient *redis.Client
	asynqClient *asynq.Client
	asynqServer *asynq.Server
	asynqMux    *asynq.ServeMux

	memCh   chan memoryJob
	memWg   sync.WaitGroup
	memCtx  context.Context
	memStop context.CancelFunc
	started atomic.Bool
	closed  atomic.Bool

	totalJobs     atomic.Int64
	completedJobs atomic.Int64
	failedJobs    atomic.Int64
	activeJobs    atomic.Int64
}

// NewRuntime initializes Redis/asynq when available and falls back to an in-memory worker pool otherwise.
func NewRuntime(cfg *config.Config, log *slog.Logger) *Runtime {
	if log == nil {
		log = slog.Default()
	}
	workers := 5
	retries := 3
	if cfg != nil && cfg.MaxConcurrentJobs > 0 {
		workers = cfg.MaxConcurrentJobs
	}
	if cfg != nil && cfg.JobRetryAttempts > 0 {
		retries = cfg.JobRetryAttempts
	}

	rt := &Runtime{
		log:           log,
		maxConcurrent: workers,
		maxRetries:    retries,
		queues: map[string]struct{}{
			QueueAutomation:     {},
			QueueBulkTaskImport: {},
		},
		handlers: make(map[string]HandlerFunc),
		memCh:    make(chan memoryJob, defaultInMemoryBuffer),
	}

	rt.initRedisOrFallback(cfg)
	return rt
}

func (r *Runtime) initRedisOrFallback(cfg *config.Config) {
	if cfg == nil || cfg.RedisHost == "" {
		r.backend = "in-memory"
		r.requestedBackend = "in-memory"
		return
	}

	r.requestedBackend = "redis"
	port := cfg.RedisPort
	if port <= 0 {
		port = 6379
	}
	addr := fmt.Sprintf("%s:%d", cfg.RedisHost, port)
	client := redis.NewClient(&redis.Options{
		Addr:        addr,
		Password:    cfg.RedisPassword,
		DB:          cfg.RedisDB,
		DialTimeout: time.Second,
	})

	pingCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if pingErr := client.Ping(pingCtx).Err(); pingErr != nil {
		_ = client.Close()
		r.backend = "in-memory"
		r.fallbackOccurred = true
		r.redisAvailable = false
		r.log.Warn("Redis unavailable; using in-memory queue fallback", slog.String("addr", addr), slog.String("error", pingErr.Error()))
		return
	}

	opt := asynq.RedisClientOpt{
		Addr:     addr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}
	r.redisClient = client
	r.asynqClient = asynq.NewClient(opt)
	r.asynqMux = asynq.NewServeMux()
	r.asynqServer = asynq.NewServer(opt, asynq.Config{
		Concurrency: r.maxConcurrent,
		Queues: map[string]int{
			QueueAutomation:     5,
			QueueBulkTaskImport: 3,
			"default":           1,
		},
	})
	r.backend = "redis"
	r.redisAvailable = true
	r.log.Info("Redis/asynq queue runtime initialized", slog.String("addr", addr))
}

// RegisterQueue records a queue name in the health registry.
func (r *Runtime) RegisterQueue(name string) {
	if name == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.queues[name] = struct{}{}
}

// RegisterHandler binds a task type processor both for asynq and the in-memory fallback pool.
func (r *Runtime) RegisterHandler(taskType string, handler HandlerFunc) {
	if taskType == "" || handler == nil {
		return
	}
	r.mu.Lock()
	r.handlers[taskType] = handler
	mux := r.asynqMux
	r.mu.Unlock()

	if mux != nil {
		mux.HandleFunc(taskType, func(ctx context.Context, t *asynq.Task) error {
			return r.executeWithMetrics(ctx, t.Type(), t.Payload(), handler)
		})
	}
}

// Start launches the asynq worker server (if Redis is connected) and in-memory fallback workers.
func (r *Runtime) Start(ctx context.Context) error {
	if !r.started.CompareAndSwap(false, true) {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	r.memCtx, r.memStop = context.WithCancel(ctx)
	for i := 0; i < r.maxConcurrent; i++ {
		r.memWg.Add(1)
		go r.runMemoryWorker()
	}

	if r.asynqServer != nil && r.asynqMux != nil {
		if err := r.asynqServer.Start(r.asynqMux); err != nil {
			r.log.Warn("Failed to start asynq server; relying on in-memory queue", slog.String("error", err.Error()))
			r.redisAvailable = false
			r.fallbackOccurred = true
			r.backend = "in-memory"
		}
	}
	return nil
}

// Close gracefully drains in-memory workers and shuts down asynq/Redis connections.
func (r *Runtime) Close() error {
	if !r.closed.CompareAndSwap(false, true) {
		return nil
	}
	if r.memStop != nil {
		r.memStop()
	}
	r.memWg.Wait()

	var firstErr error
	if r.asynqServer != nil {
		r.asynqServer.Shutdown()
	}
	if r.asynqClient != nil {
		if err := r.asynqClient.Close(); err != nil {
			firstErr = err
		}
	}
	if r.redisClient != nil {
		if err := r.redisClient.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Enqueue schedules a background job on Redis/asynq or falls back to the in-memory worker pool.
func (r *Runtime) Enqueue(ctx context.Context, queueName, taskType string, payload any) error {
	if r.closed.Load() {
		return fmt.Errorf("queue runtime is closed")
	}
	r.RegisterQueue(queueName)

	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal job payload: %w", err)
	}
	r.totalJobs.Add(1)

	if r.redisAvailable && r.asynqClient != nil {
		opts := []asynq.Option{asynq.MaxRetry(r.maxRetries)}
		if queueName != "" {
			opts = append(opts, asynq.Queue(queueName))
		}
		_, enqErr := r.asynqClient.EnqueueContext(ctx, asynq.NewTask(taskType, raw), opts...)
		if enqErr == nil {
			return nil
		}
		r.log.Warn("asynq enqueue failed; dispatching to in-memory queue", slog.String("task", taskType), slog.String("error", enqErr.Error()))
	}

	if !r.started.Load() {
		if startErr := r.Start(context.Background()); startErr != nil {
			return startErr
		}
	}

	job := memoryJob{queueName: queueName, taskType: taskType, payload: raw}
	select {
	case r.memCh <- job:
		return nil
	case <-ctx.Done():
		r.failedJobs.Add(1)
		return ctx.Err()
	}
}

func (r *Runtime) runMemoryWorker() {
	defer r.memWg.Done()
	for {
		select {
		case <-r.memCtx.Done():
			return
		case job := <-r.memCh:
			r.handleMemoryJob(job)
		}
	}
}

func (r *Runtime) handleMemoryJob(job memoryJob) {
	r.mu.RLock()
	handler := r.handlers[job.taskType]
	r.mu.RUnlock()

	if handler == nil {
		r.failedJobs.Add(1)
		r.log.Warn("No handler registered for task", slog.String("task", job.taskType))
		return
	}

	err := r.executeWithMetrics(r.memCtx, job.taskType, job.payload, handler)
	if err == nil {
		return
	}
	if job.attempt+1 < r.maxRetries && r.memCtx.Err() == nil {
		next := job
		next.attempt++
		go r.requeueWithBackoff(next)
		return
	}
	r.log.Error("Background job failed permanently", slog.String("task", job.taskType), slog.Int("attempt", job.attempt+1), slog.String("error", err.Error()))
}

func (r *Runtime) requeueWithBackoff(job memoryJob) {
	delay := time.Duration(job.attempt) * 200 * time.Millisecond
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-r.memCtx.Done():
		return
	case <-timer.C:
	}
	select {
	case r.memCh <- job:
	case <-r.memCtx.Done():
	}
}

func (r *Runtime) executeWithMetrics(ctx context.Context, taskType string, payload []byte, handler HandlerFunc) error {
	r.activeJobs.Add(1)
	defer r.activeJobs.Add(-1)

	err := handler(ctx, payload)
	if err != nil {
		r.failedJobs.Add(1)
		r.log.Debug("Queue job attempt failed", slog.String("task", taskType), slog.String("error", err.Error()))
		return err
	}
	r.completedJobs.Add(1)
	return nil
}

// Health returns queue availability, active backend, registered queues, and execution stats.
func (r *Runtime) Health() HealthResponse {
	r.mu.RLock()
	queues := make([]string, 0, len(r.queues))
	for name := range r.queues {
		queues = append(queues, name)
	}
	r.mu.RUnlock()
	slices.Sort(queues)

	total := r.totalJobs.Load()
	completed := r.completedJobs.Load()
	failed := r.failedJobs.Load()
	active := r.activeJobs.Load()

	var successRate float64
	if settled := completed + failed; settled > 0 {
		successRate = float64(completed) * 100.0 / float64(settled)
	}

	return HealthResponse{
		Backend:          r.backend,
		RequestedBackend: r.requestedBackend,
		FallbackOccurred: r.fallbackOccurred,
		RedisAvailable:   r.redisAvailable,
		Queues:           queues,
		Stats: HealthStats{
			TotalJobs:     total,
			CompletedJobs: completed,
			FailedJobs:    failed,
			ActiveJobs:    active,
			SuccessRate:   successRate,
		},
		Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
	}
}
