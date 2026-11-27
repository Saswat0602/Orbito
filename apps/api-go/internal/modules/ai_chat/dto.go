package ai_chat

import (
	"time"

	"github.com/google/uuid"
)

const iso8601UTC = "2006-01-02T15:04:05.000Z"

// ChatMessage represents a single turn in an AI conversation.
type ChatMessage struct {
	Role    string `json:"role" validate:"required,oneof=system user assistant"`
	Content string `json:"content" validate:"required"`
}

// ChatRequest matches NestJS ChatRequestDto.
type ChatRequest struct {
	Message               string        `json:"message" validate:"required"`
	History               []ChatMessage `json:"history"`
	WorkspaceID           string        `json:"workspaceId"`
	ProjectID             string        `json:"projectId"`
	SessionID             string        `json:"sessionId"`
	CurrentOrganizationID string        `json:"currentOrganizationId"`
}

// ChatResponse matches NestJS ChatResponseDto.
type ChatResponse struct {
	Message string  `json:"message"`
	Success bool    `json:"success"`
	Error   *string `json:"error,omitempty"`
}

// TestConnectionRequest matches NestJS TestConnectionDto.
type TestConnectionRequest struct {
	APIKey string `json:"apiKey"`
	Model  string `json:"model" validate:"required"`
	APIURL string `json:"apiUrl" validate:"required"`
}

// TestConnectionResponse matches NestJS TestConnectionResponseDto.
type TestConnectionResponse struct {
	Success bool    `json:"success"`
	Message *string `json:"message,omitempty"`
	Error   *string `json:"error,omitempty"`
}

// GenerateDescriptionRequest matches NestJS GenerateDescriptionDto.
type GenerateDescriptionRequest struct {
	Title    string  `json:"title" validate:"required"`
	TaskType *string `json:"taskType"`
}

// GenerateDescriptionResponse matches NestJS GenerateDescriptionResponseDto.
type GenerateDescriptionResponse struct {
	Description string  `json:"description"`
	Success     bool    `json:"success"`
	Error       *string `json:"error,omitempty"`
}

// CreateConversationRequest matches NestJS CreateConversationDto.
type CreateConversationRequest struct {
	Title     *string `json:"title"`
	SessionID *string `json:"sessionId"`
}

// RenameConversationRequest matches NestJS RenameConversationDto.
type RenameConversationRequest struct {
	Title string `json:"title" validate:"required"`
}

// UpdateMessagesRequest matches NestJS UpdateMessagesDto.
type UpdateMessagesRequest struct {
	Messages []ChatMessage `json:"messages" validate:"required"`
}

// PublicAIConfigResponse matches GET /api/ai-chat/config.
type PublicAIConfigResponse struct {
	Enabled         bool   `json:"enabled"`
	GreetingMessage string `json:"greetingMessage"`
	Tone            string `json:"tone"`
}

// MessageResponse represents a persisted chat_messages row.
type MessageResponse struct {
	ID             uuid.UUID `json:"id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	ConversationID uuid.UUID `json:"conversationId"`
	CreatedAt      string    `json:"createdAt"`
}

// ConversationResponse represents a persisted conversations row with messages.
type ConversationResponse struct {
	ID        uuid.UUID         `json:"id"`
	Title     string            `json:"title"`
	SessionID string            `json:"sessionId"`
	UserID    uuid.UUID         `json:"userId"`
	Messages  []MessageResponse `json:"messages"`
	CreatedAt string            `json:"createdAt"`
	UpdatedAt string            `json:"updatedAt"`
}

func formatTime(t time.Time) string {
	return t.UTC().Format(iso8601UTC)
}
