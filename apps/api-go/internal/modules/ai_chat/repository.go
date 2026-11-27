package ai_chat

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// Repository defines database operations for AI chat conversations and messages.
type Repository interface {
	ListConversations(ctx context.Context, userID uuid.UUID) ([]sqlc.Conversation, error)
	ListMessagesBatch(ctx context.Context, convIDs []uuid.UUID) ([]sqlc.ChatMessage, error)
	ListMessagesByConversation(ctx context.Context, convID uuid.UUID, limit int32) ([]sqlc.ChatMessage, error)
	GetConversationByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*sqlc.Conversation, error)
	GetConversationBySessionID(ctx context.Context, sessionID string) (*sqlc.Conversation, error)
	CreateConversation(ctx context.Context, id, userID uuid.UUID, title, sessionID string) (sqlc.Conversation, error)
	RenameConversation(ctx context.Context, id, userID uuid.UUID, title string) (*sqlc.Conversation, error)
	DeleteConversation(ctx context.Context, id, userID uuid.UUID) (int64, error)
	DeleteConversationBySession(ctx context.Context, sessionID string, userID uuid.UUID) (int64, error)
	InsertMessage(ctx context.Context, convID uuid.UUID, role, content string) (sqlc.ChatMessage, error)
	ReplaceMessagesTx(ctx context.Context, convID uuid.UUID, msgs []ChatMessage) error
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository creates a PostgreSQL-backed AI chat repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{
		db:      db,
		queries: sqlc.New(db.Pool()),
	}
}

func (r *pgRepository) ListConversations(ctx context.Context, userID uuid.UUID) ([]sqlc.Conversation, error) {
	return r.queries.ListConversationsByUserID(ctx, userID)
}

func (r *pgRepository) ListMessagesBatch(ctx context.Context, convIDs []uuid.UUID) ([]sqlc.ChatMessage, error) {
	if len(convIDs) == 0 {
		return nil, nil
	}
	return r.queries.ListChatMessagesBatch(ctx, convIDs)
}

func (r *pgRepository) ListMessagesByConversation(ctx context.Context, convID uuid.UUID, limit int32) ([]sqlc.ChatMessage, error) {
	return r.queries.ListChatMessagesByConversationID(ctx, sqlc.ListChatMessagesByConversationIDParams{
		ConversationID: convID,
		Limit:          limit,
	})
}

func (r *pgRepository) GetConversationByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*sqlc.Conversation, error) {
	row, err := r.queries.GetConversationByIDAndUser(ctx, sqlc.GetConversationByIDAndUserParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetConversationBySessionID(ctx context.Context, sessionID string) (*sqlc.Conversation, error) {
	row, err := r.queries.GetConversationBySessionID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) CreateConversation(ctx context.Context, id, userID uuid.UUID, title, sessionID string) (sqlc.Conversation, error) {
	row, err := r.queries.CreateConversationRecord(ctx, sqlc.CreateConversationRecordParams{
		ID:        id,
		Title:     title,
		SessionID: sessionID,
		UserID:    userID,
	})
	if err != nil {
		return sqlc.Conversation{}, apperror.Map(err)
	}
	return row, nil
}

func (r *pgRepository) RenameConversation(ctx context.Context, id, userID uuid.UUID, title string) (*sqlc.Conversation, error) {
	row, err := r.queries.RenameConversationByIDAndUser(ctx, sqlc.RenameConversationByIDAndUserParams{
		ID:     id,
		UserID: userID,
		Title:  title,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) DeleteConversation(ctx context.Context, id, userID uuid.UUID) (int64, error) {
	return r.queries.DeleteConversationByIDAndUser(ctx, sqlc.DeleteConversationByIDAndUserParams{ID: id, UserID: userID})
}

func (r *pgRepository) DeleteConversationBySession(ctx context.Context, sessionID string, userID uuid.UUID) (int64, error) {
	return r.queries.DeleteConversationBySessionIDAndUser(ctx, sqlc.DeleteConversationBySessionIDAndUserParams{SessionID: sessionID, UserID: userID})
}

func (r *pgRepository) InsertMessage(ctx context.Context, convID uuid.UUID, role, content string) (sqlc.ChatMessage, error) {
	row, err := r.queries.InsertChatMessageRecord(ctx, sqlc.InsertChatMessageRecordParams{
		ID:             uuid.New(),
		Role:           role,
		Content:        content,
		ConversationID: convID,
	})
	if err != nil {
		return sqlc.ChatMessage{}, apperror.Map(err)
	}
	return row, nil
}

func (r *pgRepository) ReplaceMessagesTx(ctx context.Context, convID uuid.UUID, msgs []ChatMessage) error {
	return r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		if err := q.DeleteMessagesByConversationID(ctx, convID); err != nil {
			return err
		}
		for _, m := range msgs {
			if _, err := q.InsertChatMessageRecord(ctx, sqlc.InsertChatMessageRecordParams{
				ID:             uuid.New(),
				Role:           m.Role,
				Content:        m.Content,
				ConversationID: convID,
			}); err != nil {
				return err
			}
		}
		return q.TouchConversationUpdatedAt(ctx, convID)
	})
}
