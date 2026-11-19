package email

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/orbito/api-go/internal/platform/mailer"
)

type mockMailer struct {
	mu       sync.Mutex
	sentMsgs []mailer.Message
}

func (m *mockMailer) Send(_ context.Context, msg mailer.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentMsgs = append(m.sentMsgs, msg)
	return nil
}

func (m *mockMailer) Verify(_ context.Context) error {
	return nil
}

func TestEmailServiceAndQueue(t *testing.T) {
	mock := &mockMailer{}
	svc := NewService(mock, 2)
	defer svc.Close()

	ctx := context.Background()

	// Send single email
	err := svc.SendEmail(ctx, SendEmailDTO{
		To:       "user@example.com",
		Subject:  "Welcome to Orbito",
		Template: "task-assigned",
		Data:     map[string]any{"task": map[string]any{"title": "Initial Task"}},
	})
	if err != nil {
		t.Fatalf("Failed to queue email: %v", err)
	}

	// Send bulk email
	err = svc.SendBulkEmail(ctx, BulkEmailDTO{
		Recipients: []string{"a@example.com", "b@example.com"},
		Subject:    "Bulk Notice",
		Template:   "task-assigned",
		Data:       map[string]any{"task": map[string]any{"title": "Sprint 1"}},
	})
	if err != nil {
		t.Fatalf("Failed to queue bulk emails: %v", err)
	}

	// Wait for queue processing
	time.Sleep(100 * time.Millisecond)

	stats := svc.GetQueueStats()
	if stats.Completed < 3 {
		t.Fatalf("Expected at least 3 completed emails, got stats: %+v", stats)
	}

	mock.mu.Lock()
	defer mock.mu.Unlock()
	if len(mock.sentMsgs) < 3 {
		t.Fatalf("Expected 3 sent messages, got %d", len(mock.sentMsgs))
	}
}

func TestEmailHTTPHandler(t *testing.T) {
	mock := &mockMailer{}
	svc := NewService(mock, 1)
	defer svc.Close()

	handler := NewHandler(svc)
	router := Routes(handler, nil)

	t.Run("POST /send queues email successfully", func(t *testing.T) {
		body, _ := json.Marshal(SendEmailDTO{
			To:       "john@domain.com",
			Subject:  "Hello John",
			Template: "task-assigned",
			Data:     map[string]any{"message": "Test"},
		})

		req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Fatalf("Expected 202 Accepted, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("POST /send validates email format", func(t *testing.T) {
		body, _ := json.Marshal(SendEmailDTO{
			To:       "invalid-email-string",
			Subject:  "Hello",
			Template: "task-assigned",
			Data:     map[string]any{},
		})

		req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Expected 400 Bad Request, got %d", rec.Code)
		}
	})

	t.Run("POST /send-bulk queues multiple emails", func(t *testing.T) {
		body, _ := json.Marshal(BulkEmailDTO{
			Recipients: []string{"one@domain.com", "two@domain.com"},
			Subject:    "Bulk Update",
			Template:   "task-assigned",
			Data:       map[string]any{"message": "Bulk"},
		})

		req := httptest.NewRequest(http.MethodPost, "/send-bulk", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Fatalf("Expected 202 Accepted, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("GET /queue/stats returns queue counters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/queue/stats", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d", rec.Code)
		}

		var stats QueueStatsResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &stats); err != nil {
			t.Fatalf("Failed to parse queue stats response: %v", err)
		}
	})
}
