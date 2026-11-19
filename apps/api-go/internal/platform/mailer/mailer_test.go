package mailer

import (
	"context"
	"strings"
	"testing"
)

func TestMailerSimulation(t *testing.T) {
	cfg := Config{
		Host: "smtp.example.com",
		Port: 587,
		User: "user@example.com",
		Pass: "",
		From: "noreply@orbito.io",
	}

	m := New(cfg)

	// Verify should succeed in simulation mode
	if err := m.Verify(context.Background()); err != nil {
		t.Fatalf("Expected simulation Verify to succeed, got %v", err)
	}

	// Send should succeed in simulation mode
	msg := Message{
		To:      []string{"user@test.local"},
		Subject: "Test Email",
		HTML:    "<h1>Hello</h1>",
	}
	if err := m.Send(context.Background(), msg); err != nil {
		t.Fatalf("Expected simulation Send to succeed, got %v", err)
	}
}

func TestRenderTemplate(t *testing.T) {
	t.Run("task-assigned template", func(t *testing.T) {
		rendered := RenderTemplate("task-assigned", map[string]any{
			"task": map[string]any{
				"title": "Fix critical login bug",
			},
			"assigner": map[string]any{
				"name": "Alex",
			},
		})
		if !strings.Contains(rendered, "Fix critical login bug") {
			t.Fatal("Rendered email missing task title")
		}
		if !strings.Contains(rendered, "Alex") {
			t.Fatal("Rendered email missing assigner name")
		}
	})

	t.Run("password-reset template", func(t *testing.T) {
		rendered := RenderTemplate("password-reset", map[string]any{
			"resetLink": "https://orbito.io/reset?token=abc123xyz",
		})
		if !strings.Contains(rendered, "https://orbito.io/reset?token=abc123xyz") {
			t.Fatal("Rendered email missing reset link")
		}
	})

	t.Run("invitation template", func(t *testing.T) {
		rendered := RenderTemplate("send-invitation", map[string]any{
			"workspace": map[string]any{
				"name": "Acme Corp",
			},
			"inviteLink": "https://orbito.io/invite/join?token=123",
		})
		if !strings.Contains(rendered, "Acme Corp") {
			t.Fatal("Rendered email missing workspace name")
		}
	})
}
