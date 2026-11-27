package ai_chat

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/platform/safehttp"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// SettingsReader reads per-user or global configuration keys (decrypted).
type SettingsReader interface {
	Get(ctx context.Context, key string, userID *uuid.UUID, defaultValue *string) (*string, error)
}

// Service implements AI assistant chat, description generation, connection test, and conversations.
type Service interface {
	GetPublicConfig(ctx context.Context, userID uuid.UUID) (*PublicAIConfigResponse, error)
	GetConversations(ctx context.Context, userID uuid.UUID) ([]ConversationResponse, error)
	CreateConversation(ctx context.Context, userID uuid.UUID, req CreateConversationRequest) (*ConversationResponse, error)
	RenameConversation(ctx context.Context, userID, id uuid.UUID, req RenameConversationRequest) (*ConversationResponse, error)
	DeleteConversation(ctx context.Context, userID, id uuid.UUID) error
	UpdateMessages(ctx context.Context, userID, id uuid.UUID, req UpdateMessagesRequest) error
	Chat(ctx context.Context, userID uuid.UUID, req ChatRequest) (*ChatResponse, error)
	TestConnection(ctx context.Context, req TestConnectionRequest) (*TestConnectionResponse, error)
	GenerateDescription(ctx context.Context, userID uuid.UUID, req GenerateDescriptionRequest) (*GenerateDescriptionResponse, error)
	ClearContext(ctx context.Context, userID uuid.UUID, sessionID string) error
}

type service struct {
	repo          Repository
	settings      SettingsReader
	allowlist     []string
	allowPrivate  bool
	appURL        string
	resolveDestFn func(ctx context.Context, rawURL string, opts safehttp.Options) (*safehttp.Destination, error)
	log           *slog.Logger
}

// NewService constructs an AI chat service.
func NewService(repo Repository, settings SettingsReader, cfg *config.Config, log *slog.Logger) Service {
	if log == nil {
		log = slog.Default()
	}
	var rawAllow []string
	var allowPriv bool
	var appURL string
	if cfg != nil {
		rawAllow = cfg.AIAllowedHosts
		allowPriv = cfg.AIAllowPrivateEndpoints
		appURL = cfg.FrontendURL
	}
	allowlist := safehttp.ParseHostAllowlist(rawAllow, []string{"api.openai.com", "*.anthropic.com", "openrouter.ai"})
	return &service{
		repo:          repo,
		settings:      settings,
		allowlist:     allowlist,
		allowPrivate:  allowPriv,
		appURL:        appURL,
		resolveDestFn: safehttp.ResolveSafeDestination,
		log:           log,
	}
}

func (s *service) getSetting(ctx context.Context, key string, userID *uuid.UUID) string {
	if s.settings == nil {
		return ""
	}
	val, err := s.settings.Get(ctx, key, userID, nil)
	if err != nil || val == nil {
		return ""
	}
	return *val
}

func (s *service) GetPublicConfig(ctx context.Context, userID uuid.UUID) (*PublicAIConfigResponse, error) {
	enabled := s.getSetting(ctx, "ai_enabled", &userID) == "true"
	greeting := s.getSetting(ctx, "ai_greeting_message", &userID)
	if greeting == "" {
		greeting = "How can I help you today?"
	}
	tone := s.getSetting(ctx, "ai_tone", &userID)
	if tone == "" {
		tone = "Professional"
	}
	return &PublicAIConfigResponse{
		Enabled:         enabled,
		GreetingMessage: greeting,
		Tone:            tone,
	}, nil
}

func (s *service) GetConversations(ctx context.Context, userID uuid.UUID) ([]ConversationResponse, error) {
	convs, err := s.repo.ListConversations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	ids := make([]uuid.UUID, 0, len(convs))
	for i := range convs {
		ids = append(ids, convs[i].ID)
	}
	msgs, _ := s.repo.ListMessagesBatch(ctx, ids)
	msgMap := make(map[uuid.UUID][]MessageResponse)
	for i := range msgs {
		m := msgs[i]
		bucket := msgMap[m.ConversationID]
		if len(bucket) < 40 {
			msgMap[m.ConversationID] = append(bucket, mapMessage(&m))
		}
	}

	out := make([]ConversationResponse, 0, len(convs))
	for i := range convs {
		c := &convs[i]
		cMsgs := msgMap[c.ID]
		if cMsgs == nil {
			cMsgs = []MessageResponse{}
		}
		out = append(out, mapConversation(c, cMsgs))
	}
	return out, nil
}

func (s *service) CreateConversation(ctx context.Context, userID uuid.UUID, req CreateConversationRequest) (*ConversationResponse, error) {
	title := "New Chat"
	if req.Title != nil && strings.TrimSpace(*req.Title) != "" {
		title = strings.TrimSpace(*req.Title)
	}
	sessionID := ""
	if req.SessionID != nil && strings.TrimSpace(*req.SessionID) != "" {
		sessionID = strings.TrimSpace(*req.SessionID)
	} else {
		var buf [16]byte
		_, _ = rand.Read(buf[:])
		sessionID = fmt.Sprintf("session_%d_%s", time.Now().UnixMilli(), hex.EncodeToString(buf[:]))
	}

	row, err := s.repo.CreateConversation(ctx, uuid.New(), userID, title, sessionID)
	if err != nil {
		return nil, err
	}
	res := mapConversation(&row, []MessageResponse{})
	return &res, nil
}

func (s *service) RenameConversation(ctx context.Context, userID, id uuid.UUID, req RenameConversationRequest) (*ConversationResponse, error) {
	row, err := s.repo.RenameConversation(ctx, id, userID, strings.TrimSpace(req.Title))
	if err != nil {
		return nil, fmt.Errorf("rename conversation: %w", err)
	}
	if row == nil {
		return nil, apperror.NotFound("Conversation not found")
	}
	msgs, _ := s.repo.ListMessagesByConversation(ctx, row.ID, 40)
	mapped := make([]MessageResponse, 0, len(msgs))
	for i := range msgs {
		mapped = append(mapped, mapMessage(&msgs[i]))
	}
	res := mapConversation(row, mapped)
	return &res, nil
}

func (s *service) DeleteConversation(ctx context.Context, userID, id uuid.UUID) error {
	affected, err := s.repo.DeleteConversation(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("delete conversation: %w", err)
	}
	if affected == 0 {
		return apperror.NotFound("Conversation not found")
	}
	return nil
}

func (s *service) UpdateMessages(ctx context.Context, userID, id uuid.UUID, req UpdateMessagesRequest) error {
	conv, err := s.repo.GetConversationByIDAndUser(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("find conversation: %w", err)
	}
	if conv == nil {
		return apperror.BadRequest("Conversation not found")
	}
	return s.repo.ReplaceMessagesTx(ctx, conv.ID, req.Messages)
}

func (s *service) ClearContext(ctx context.Context, userID uuid.UUID, sessionID string) error {
	_, err := s.repo.DeleteConversationBySession(ctx, sessionID, userID)
	return err
}

func (s *service) Chat(ctx context.Context, userID uuid.UUID, req ChatRequest) (*ChatResponse, error) {
	if s.getSetting(ctx, "ai_enabled", &userID) != "true" {
		return nil, apperror.BadRequest("AI chat is currently disabled. Please enable it in settings.")
	}
	rawAPIURL := s.getSetting(ctx, "ai_api_url", &userID)
	if rawAPIURL == "" {
		return nil, apperror.BadRequest("AI API URL not configured. Please set the API URL in settings.")
	}
	apiKey := s.getSetting(ctx, "ai_api_key", &userID)
	model := s.getSetting(ctx, "ai_model", &userID)
	if model == "" {
		model = "gpt-4o-mini"
	}

	dest, provider, err := s.resolveProvider(ctx, rawAPIURL)
	if err != nil {
		return nil, err
	}
	if apiKey == "" && provider != "ollama" {
		return nil, apperror.BadRequest("AI API key not configured. Please set it in settings.")
	}

	conv, dbHistory, cErr := s.resolveOrCreateSession(ctx, userID, req.SessionID)
	if cErr != nil {
		return nil, cErr
	}

	history := req.History
	if len(history) == 0 {
		history = dbHistory
	}
	tone := s.getSetting(ctx, "ai_tone", &userID)
	customSys := s.getSetting(ctx, "ai_system_prompt", &userID)

	msgs := []ChatMessage{{Role: "system", Content: BuildSystemPrompt(tone, customSys)}}
	msgs = append(msgs, history...)
	userMsg := req.Message
	if strings.Contains(userMsg, "Task:") && strings.Contains(userMsg, "Current URL:") {
		userMsg += "\n\n" + EnhancePromptWithContext(userMsg, "")
		if guide := GetAutomationPrompt(userMsg); guide != "" {
			userMsg += "\n\n" + guide
		}
	}
	msgs = append(msgs, ChatMessage{Role: "user", Content: userMsg})

	normalized := NormalizeMessages(msgs)
	reply, callErr := s.callProvider(ctx, dest, provider, apiKey, model, normalized, 1500, 0.7, true)
	if callErr != nil {
		errMsg := callErr.Error()
		return &ChatResponse{Message: "", Success: false, Error: &errMsg}, nil
	}

	if conv != nil {
		_, _ = s.repo.InsertMessage(ctx, conv.ID, "user", req.Message)
		_, _ = s.repo.InsertMessage(ctx, conv.ID, "assistant", reply)
		if conv.Title == "New Chat" && len(strings.TrimSpace(req.Message)) > 0 {
			title := strings.TrimSpace(req.Message)
			if len(title) > 40 {
				title = title[:40] + "..."
			}
			_, _ = s.repo.RenameConversation(ctx, conv.ID, userID, title)
		}
	}
	return &ChatResponse{Message: reply, Success: true}, nil
}

func (s *service) TestConnection(ctx context.Context, req TestConnectionRequest) (*TestConnectionResponse, error) {
	dest, provider, err := s.resolveProvider(ctx, req.APIURL)
	if err != nil {
		msg := err.Error()
		return &TestConnectionResponse{Success: false, Error: &msg}, nil
	}
	if strings.TrimSpace(req.APIKey) == "" && provider != "ollama" {
		msg := "API key is required for this provider."
		return &TestConnectionResponse{Success: false, Error: &msg}, nil
	}

	msgs := []ChatMessage{{Role: "user", Content: `Hello, this is a connection test. Please respond with "Connection successful."`}}
	reply, callErr := s.callProvider(ctx, dest, provider, req.APIKey, req.Model, msgs, 50, 0.1, false)
	if callErr != nil {
		msg := callErr.Error()
		return &TestConnectionResponse{Success: false, Error: &msg}, nil
	}
	if strings.TrimSpace(reply) == "" {
		msg := "Received empty response from AI provider. Please check your configuration."
		return &TestConnectionResponse{Success: false, Error: &msg}, nil
	}
	okMsg := "Connection successful! Your AI configuration is working correctly."
	return &TestConnectionResponse{Success: true, Message: &okMsg}, nil
}

func (s *service) GenerateDescription(ctx context.Context, userID uuid.UUID, req GenerateDescriptionRequest) (*GenerateDescriptionResponse, error) {
	if s.getSetting(ctx, "ai_enabled", &userID) != "true" {
		errMsg := "AI is not enabled."
		return &GenerateDescriptionResponse{Success: false, Error: &errMsg}, nil
	}
	rawAPIURL := s.getSetting(ctx, "ai_api_url", &userID)
	model := s.getSetting(ctx, "ai_model", &userID)
	apiKey := s.getSetting(ctx, "ai_api_key", &userID)
	if rawAPIURL == "" || model == "" {
		errMsg := "AI not configured."
		return &GenerateDescriptionResponse{Success: false, Error: &errMsg}, nil
	}
	dest, provider, err := s.resolveProvider(ctx, rawAPIURL)
	if err != nil {
		errMsg := err.Error()
		return &GenerateDescriptionResponse{Success: false, Error: &errMsg}, nil
	}
	if apiKey == "" && provider != "ollama" {
		errMsg := "AI API key not configured."
		return &GenerateDescriptionResponse{Success: false, Error: &errMsg}, nil
	}

	taskType := "task"
	if req.TaskType != nil && *req.TaskType != "" {
		taskType = strings.ToLower(*req.TaskType)
	}
	msgs := []ChatMessage{
		{Role: "system", Content: "Generate a concise 2-4 sentence Markdown description for a project management task with summary and acceptance criteria."},
		{Role: "user", Content: fmt.Sprintf("Generate a description for this %s:\nTitle: %q", taskType, req.Title)},
	}
	reply, callErr := s.callProvider(ctx, dest, provider, apiKey, model, msgs, 400, 0.5, false)
	if callErr != nil {
		errMsg := callErr.Error()
		return &GenerateDescriptionResponse{Success: false, Error: &errMsg}, nil
	}
	return &GenerateDescriptionResponse{Description: strings.TrimSpace(reply), Success: true}, nil
}

func (s *service) resolveProvider(ctx context.Context, rawURL string) (*safehttp.Destination, string, error) {
	dest, err := s.resolveDestFn(ctx, rawURL, safehttp.Options{
		Allowlist:    s.allowlist,
		AllowPrivate: s.allowPrivate,
		Timeout:      120 * time.Second,
	})
	if err != nil {
		return nil, "", apperror.BadRequest(err.Error())
	}
	return dest, DetectProvider(dest.URL), nil
}

func (s *service) resolveOrCreateSession(ctx context.Context, userID uuid.UUID, sessionID string) (*sqlc.Conversation, []ChatMessage, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, nil, nil
	}
	existing, err := s.repo.GetConversationBySessionID(ctx, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("lookup conversation session: %w", err)
	}
	if existing != nil {
		if existing.UserID != userID {
			return nil, nil, apperror.NotFound("Conversation not found")
		}
		msgs, _ := s.repo.ListMessagesByConversation(ctx, existing.ID, 40)
		hist := make([]ChatMessage, 0, len(msgs))
		for i := range msgs {
			hist = append(hist, ChatMessage{Role: msgs[i].Role, Content: msgs[i].Content})
		}
		return existing, hist, nil
	}
	created, err := s.repo.CreateConversation(ctx, uuid.New(), userID, "New Chat", sessionID)
	if err != nil {
		return nil, nil, err
	}
	return &created, nil, nil
}

func (s *service) callProvider(
	ctx context.Context,
	dest *safehttp.Destination,
	provider, apiKey, model string,
	messages []ChatMessage,
	maxTokens int,
	temperature float64,
	samplingExtras bool,
) (string, error) {
	reqURL, headers, payload, err := BuildProviderRequest(provider, dest.URL, apiKey, model, s.appURL, messages, maxTokens, temperature, samplingExtras)
	if err != nil {
		return "", err
	}
	rawBody, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(rawBody))
	if err != nil {
		return "", err
	}
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}
	if parsed, pErr := url.Parse(reqURL); pErr == nil && parsed.Host != "" {
		httpReq.Host = parsed.Host
	}

	resp, err := dest.Client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed reaching AI provider: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBytes, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("%s", mapProviderHTTPError(resp.StatusCode, respBytes, provider))
	}

	var decoded map[string]any
	if err := json.Unmarshal(respBytes, &decoded); err != nil {
		return "", fmt.Errorf("invalid AI response JSON: %w", err)
	}
	return ParseProviderResponse(provider, decoded), nil
}

func mapProviderHTTPError(status int, body []byte, provider string) string {
	switch status {
	case http.StatusUnauthorized:
		return "Invalid API key. Please check your API key and try again."
	case http.StatusTooManyRequests:
		return "Rate limit exceeded. Please try again in a moment."
	case http.StatusPaymentRequired:
		return "Insufficient credits. Please check your account balance."
	case http.StatusNotFound:
		return "Model not found. Please check the model name and try again."
	default:
		return ReadAPIError(status, body, provider)
	}
}

func mapMessage(m *sqlc.ChatMessage) MessageResponse {
	return MessageResponse{
		ID:             m.ID,
		Role:           m.Role,
		Content:        m.Content,
		ConversationID: m.ConversationID,
		CreatedAt:      formatTime(m.CreatedAt.Time),
	}
}

func mapConversation(c *sqlc.Conversation, msgs []MessageResponse) ConversationResponse {
	return ConversationResponse{
		ID:        c.ID,
		Title:     c.Title,
		SessionID: c.SessionID,
		UserID:    c.UserID,
		Messages:  msgs,
		CreatedAt: formatTime(c.CreatedAt.Time),
		UpdatedAt: formatTime(c.UpdatedAt.Time),
	}
}
