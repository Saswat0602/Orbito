package email

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/orbito/api-go/internal/platform/mailer"
)

// Service coordinates email sending and background queue processing.
type Service struct {
	mailer    mailer.Mailer
	jobQueue  chan emailJob
	waiting   int64
	active    int64
	completed int64
	failed    int64
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

type emailJob struct {
	to       string
	subject  string
	template string
	data     map[string]any
	delay    time.Duration
}

// NewService constructs a new email service and spawns worker routines.
func NewService(m mailer.Mailer, workers int) *Service {
	if workers <= 0 {
		workers = 2
	}

	s := &Service{
		mailer:   m,
		jobQueue: make(chan emailJob, 1000),
		stopChan: make(chan struct{}),
	}

	for i := 0; i < workers; i++ {
		s.wg.Add(1)
		go s.worker()
	}

	return s
}

// Close gracefully terminates the email worker pool.
func (s *Service) Close() {
	close(s.stopChan)
	s.wg.Wait()
}

func (s *Service) worker() {
	defer s.wg.Done()
	for {
		select {
		case <-s.stopChan:
			return
		case job, ok := <-s.jobQueue:
			if !ok {
				return
			}
			atomic.AddInt64(&s.waiting, -1)
			atomic.AddInt64(&s.active, 1)

			if job.delay > 0 {
				select {
				case <-time.After(job.delay):
				case <-s.stopChan:
					atomic.AddInt64(&s.active, -1)
					return
				}
			}

			htmlBody := mailer.RenderTemplate(job.template, job.data)
			msg := mailer.Message{
				To:      []string{job.to},
				Subject: job.subject,
				HTML:    htmlBody,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			err := s.mailer.Send(ctx, msg)
			cancel()

			atomic.AddInt64(&s.active, -1)
			if err != nil {
				atomic.AddInt64(&s.failed, 1)
			} else {
				atomic.AddInt64(&s.completed, 1)
			}
		}
	}
}

// SendEmail queues a single email for delivery.
func (s *Service) SendEmail(_ context.Context, dto SendEmailDTO) error {
	var delay time.Duration
	if dto.Delay > 0 {
		delay = time.Duration(dto.Delay) * time.Millisecond
	}

	job := emailJob{
		to:       dto.To,
		subject:  dto.Subject,
		template: dto.Template,
		data:     dto.Data,
		delay:    delay,
	}

	atomic.AddInt64(&s.waiting, 1)
	select {
	case s.jobQueue <- job:
		return nil
	default:
		atomic.AddInt64(&s.waiting, -1)
		return fmt.Errorf("email queue is full")
	}
}

// SendBulkEmail queues multiple emails for delivery.
func (s *Service) SendBulkEmail(ctx context.Context, dto BulkEmailDTO) error {
	for _, recipient := range dto.Recipients {
		err := s.SendEmail(ctx, SendEmailDTO{
			To:       recipient,
			Subject:  dto.Subject,
			Template: dto.Template,
			Data:     dto.Data,
			Priority: dto.Priority,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// GetQueueStats returns current in-flight and processed email counts.
func (s *Service) GetQueueStats() QueueStatsResponse {
	return QueueStatsResponse{
		Waiting:   int(atomic.LoadInt64(&s.waiting)),
		Active:    int(atomic.LoadInt64(&s.active)),
		Completed: int(atomic.LoadInt64(&s.completed)),
		Failed:    int(atomic.LoadInt64(&s.failed)),
	}
}
