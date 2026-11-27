package ai_chat

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/platform/safehttp"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRepo struct {
	convs map[uuid.UUID]sqlc.Conversation
	msgs  []sqlc.ChatMessage
}

func (s *stubRepo) ListConversations(_ context.Context, userID uuid.UUID) ([]sqlc.Conversation, error) {
	var out []sqlc.Conversation
	for _, c := range s.convs {
		if c.UserID == userID {
			out = append(out, c)
		}
	}
	return out, nil
}
func (s *stubRepo) ListMessagesBatch(_ context.Context, _ []uuid.UUID) ([]sqlc.ChatMessage, error) {
	return s.msgs, nil
}
func (s *stubRepo) ListMessagesByConversation(_ context.Context, convID uuid.UUID, _ int32) ([]sqlc.ChatMessage, error) {
	var out []sqlc.ChatMessage
	for _, m := range s.msgs {
		if m.ConversationID == convID {
			out = append(out, m)
		}
	}
	return out, nil
}
func (s *stubRepo) GetConversationByIDAndUser(_ context.Context, id, userID uuid.UUID) (*sqlc.Conversation, error) {
	c, ok := s.convs[id]
	if !ok || c.UserID != userID {
		return nil, nil
	}
	return &c, nil
}
func (s *stubRepo) GetConversationBySessionID(_ context.Context, sessionID string) (*sqlc.Conversation, error) {
	for _, c := range s.convs {
		if c.SessionID == sessionID {
			copyC := c
			return &copyC, nil
		}
	}
	return nil, nil
}
func (s *stubRepo) CreateConversation(_ context.Context, id, userID uuid.UUID, title, sessionID string) (sqlc.Conversation, error) {
	now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
	c := sqlc.Conversation{ID: id, Title: title, SessionID: sessionID, UserID: userID, CreatedAt: now, UpdatedAt: now}
	s.convs[id] = c
	return c, nil
}
func (s *stubRepo) RenameConversation(_ context.Context, id, userID uuid.UUID, title string) (*sqlc.Conversation, error) {
	c, ok := s.convs[id]
	if !ok || c.UserID != userID {
		return nil, nil
	}
	c.Title = title
	s.convs[id] = c
	return &c, nil
}
func (s *stubRepo) DeleteConversation(_ context.Context, id, userID uuid.UUID) (int64, error) {
	c, ok := s.convs[id]
	if !ok || c.UserID != userID {
		return 0, nil
	}
	delete(s.convs, id)
	return 1, nil
}
func (s *stubRepo) DeleteConversationBySession(_ context.Context, sessionID string, userID uuid.UUID) (int64, error) {
	for id, c := range s.convs {
		if c.SessionID == sessionID && c.UserID == userID {
			delete(s.convs, id)
			return 1, nil
		}
	}
	return 0, nil
}
func (s *stubRepo) InsertMessage(_ context.Context, convID uuid.UUID, role, content string) (sqlc.ChatMessage, error) {
	m := sqlc.ChatMessage{
		ID: uuid.New(), ConversationID: convID, Role: role, Content: content,
		CreatedAt: pgtype.Timestamp{Time: time.Now().UTC(), Valid: true},
	}
	s.msgs = append(s.msgs, m)
	return m, nil
}
func (s *stubRepo) ReplaceMessagesTx(_ context.Context, convID uuid.UUID, msgs []ChatMessage) error {
	for _, m := range msgs {
		_, _ = s.InsertMessage(context.Background(), convID, m.Role, m.Content)
	}
	return nil
}

type stubSettings map[string]string

func (s stubSettings) Get(_ context.Context, key string, _ *uuid.UUID, def *string) (*string, error) {
	if v, ok := s[key]; ok {
		copyV := v
		return &copyV, nil
	}
	return def, nil
}

func TestNormalizeMessagesAndProviders(t *testing.T) {
	norm := NormalizeMessages([]ChatMessage{
		{Role: "system", Content: "sys"},
		{Role: "user", Content: "   "},
		{Role: "user", Content: "hello"},
		{Role: "user", Content: "world"},
		{Role: "assistant", Content: "hi"},
	})
	require.Len(t, norm, 3)
	assert.Equal(t, "hello\n\nworld", norm[1].Content)
	assert.Equal(t, "anthropic", DetectProvider("https://api.anthropic.com/v1"))
	assert.Equal(t, "google", DetectProvider("https://generativelanguage.googleapis.com/v1beta"))
}

func TestAiChatHandlerAndTestConnection(t *testing.T) {
	llmSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"Connection successful."}}]}`))
	}))
	defer llmSrv.Close()

	repo := &stubRepo{convs: make(map[uuid.UUID]sqlc.Conversation)}
	settings := stubSettings{
		"ai_enabled":          "true",
		"ai_api_url":          llmSrv.URL,
		"ai_api_key":          "sk-test",
		"ai_model":            "gpt-4o-mini",
		"ai_greeting_message": "Welcome!",
	}

	svc := NewService(repo, settings, &config.Config{AIAllowedHosts: []string{"127.0.0.1"}, AIAllowPrivateEndpoints: true}, nil).(*service)
	svc.resolveDestFn = func(_ context.Context, rawURL string, opts safehttp.Options) (*safehttp.Destination, error) {
		return safehttp.ResolveSafeDestination(context.Background(), rawURL, opts)
	}

	h := NewHandler(svc)
	userID := uuid.New()
	r := chi.NewRouter()
	r.Mount("/ai-chat", Routes(h, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(w, req.WithContext(ctxhelper.WithUserID(req.Context(), userID.String())))
		})
	}))

	cfgReq := httptest.NewRequest(http.MethodGet, "/ai-chat/config", nil)
	cfgRec := httptest.NewRecorder()
	r.ServeHTTP(cfgRec, cfgReq)
	assert.Equal(t, http.StatusOK, cfgRec.Code)
	assert.Contains(t, cfgRec.Body.String(), `"greetingMessage":"Welcome!"`)

	tcBody, _ := json.Marshal(TestConnectionRequest{APIKey: "sk-test", Model: "gpt-4o-mini", APIURL: llmSrv.URL})
	tcReq := httptest.NewRequest(http.MethodPost, "/ai-chat/test-connection", bytes.NewReader(tcBody))
	tcReq.Header.Set("Content-Type", "application/json")
	tcRec := httptest.NewRecorder()
	r.ServeHTTP(tcRec, tcReq)
	assert.Equal(t, http.StatusOK, tcRec.Code)
	assert.Contains(t, tcRec.Body.String(), `"success":true`)

	chatBody, _ := json.Marshal(ChatRequest{Message: "How do I create a task?", SessionID: "sess-1"})
	chatReq := httptest.NewRequest(http.MethodPost, "/ai-chat/chat", bytes.NewReader(chatBody))
	chatReq.Header.Set("Content-Type", "application/json")
	chatRec := httptest.NewRecorder()
	r.ServeHTTP(chatRec, chatReq)
	assert.Equal(t, http.StatusOK, chatRec.Code)
	assert.Contains(t, chatRec.Body.String(), `"Connection successful."`)
	assert.Len(t, repo.msgs, 2)
}
