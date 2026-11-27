package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/modules/email"
	"github.com/robfig/cron/v3"
)

// NotificationCreator abstracts creating in-app notifications from cron jobs.
type NotificationCreator interface {
	NotifyTaskDueSoon(ctx context.Context, taskID uuid.UUID) error
}

// EmailSender abstracts queuing email notifications from cron jobs.
type EmailSender interface {
	SendEmail(ctx context.Context, dto email.SendEmailDTO) error
}

type scheduledJob struct {
	spec string
	name string
	fn   func(context.Context)
}

// Scheduler registers and executes Orbito background cron schedules.
type Scheduler struct {
	cron      *cron.Cron
	repo      Repository
	notifier  NotificationCreator
	emailSvc  EmailSender
	log       *slog.Logger
	nowFunc   func() time.Time
	mu        sync.Mutex
	started   bool
	extraJobs []scheduledJob
	baseCtx   context.Context
	cancelCtx context.CancelFunc
}

// RegisterJob appends a periodic background job before Start is called.
func (s *Scheduler) RegisterJob(spec, name string, fn func(context.Context)) {
	if fn == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.extraJobs = append(s.extraJobs, scheduledJob{spec: spec, name: name, fn: fn})
}

// New constructs a Scheduler with standard 5-field cron parsing.
func New(repo Repository, notifier NotificationCreator, emailSvc EmailSender, log *slog.Logger) *Scheduler {
	if log == nil {
		log = slog.Default()
	}
	return &Scheduler{
		cron:     cron.New(),
		repo:     repo,
		notifier: notifier,
		emailSvc: emailSvc,
		log:      log,
		nowFunc:  time.Now,
	}
}

// Start registers all 7 Orbito cron jobs and starts the cron loop.
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return nil
	}
	s.baseCtx, s.cancelCtx = context.WithCancel(context.Background())

	jobs := []scheduledJob{
		{"0 * * * *", "due-date-reminders", s.CheckDueDateReminders},
		{"0 9 * * *", "overdue-tasks", s.CheckOverdueTasks},
		{"0 9 * * 1", "weekly-summaries", s.SendWeeklySummaries},
		{"0 0 * * *", "cleanup-old-data", s.CleanupOldData},
		{"0 */6 * * *", "update-sprint-statuses", s.UpdateSprintStatuses},
		{"0 0 * * 0", "generate-analytics", s.GenerateAnalytics},
		{"0 0 * * *", "generate-recurring-tasks", s.GenerateUpcomingRecurringTasks},
	}
	jobs = append(jobs, s.extraJobs...)

	for _, j := range jobs {
		jobFn := j.fn
		jobName := j.name
		if _, err := s.cron.AddFunc(j.spec, func() {
			ctx, cancel := context.WithTimeout(s.baseCtx, 5*time.Minute)
			defer cancel()
			s.log.Debug("Executing scheduled job", slog.String("job", jobName))
			jobFn(ctx)
		}); err != nil {
			return err
		}
	}

	s.cron.Start()
	s.started = true
	s.log.Info("Scheduler started", slog.Int("jobs", len(jobs)))
	return nil
}

// Stop halts the cron runner and waits for active jobs to finish.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return
	}
	s.started = false
	cancel := s.cancelCtx
	c := s.cron
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	doneCtx := c.Stop()
	<-doneCtx.Done()
}
