package inbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/modules/task_ranks"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// Repository defines persistence operations for project inboxes, accounts, messages, and rules.
type Repository interface {
	GetProjectByID(ctx context.Context, projectID uuid.UUID) (*sqlc.GetProjectByIDForInboxRow, error)
	GetInboxByProjectID(ctx context.Context, projectID uuid.UUID) (*sqlc.GetProjectInboxByProjectIDRow, error)
	FindInboxByEmail(ctx context.Context, email string) (*sqlc.GetProjectInboxByEmailAddressRow, error)
	CreateInbox(ctx context.Context, arg sqlc.CreateProjectInboxRecordParams) (sqlc.ProjectInbox, error)
	UpdateInbox(ctx context.Context, arg sqlc.UpdateProjectInboxRecordParams) (sqlc.ProjectInbox, error)
	UpdateInboxEmailAddress(ctx context.Context, inboxID uuid.UUID, email string) error

	GetEmailAccountByInboxID(ctx context.Context, inboxID uuid.UUID) (*sqlc.EmailAccount, error)
	GetEmailAccountByID(ctx context.Context, accountID uuid.UUID) (*sqlc.EmailAccount, error)
	UpsertEmailAccount(ctx context.Context, arg sqlc.UpsertEmailAccountForInboxParams) (sqlc.EmailAccount, error)
	ListSyncEnabledAccounts(ctx context.Context) ([]sqlc.ListSyncEnabledEmailAccountsRow, error)
	GetAccountSyncDetailsByProjectID(ctx context.Context, projectID uuid.UUID) (*sqlc.GetEmailAccountSyncDetailsByProjectIDRow, error)
	UpdateEmailAccountSyncState(ctx context.Context, accountID uuid.UUID, lastSync time.Time, lastErr *string, uidNext *int32, uidValidity *string) error
	CreateSyncLog(ctx context.Context, arg sqlc.CreateEmailSyncLogParams) error

	ListMessages(ctx context.Context, arg sqlc.ListInboxMessagesFilteredParams) ([]sqlc.ListInboxMessagesFilteredRow, error)
	ListMessageAttachmentsBatch(ctx context.Context, messageIDs []uuid.UUID) ([]sqlc.ListMessageAttachmentsBatchRow, error)
	GetMessageDetailByID(ctx context.Context, messageID uuid.UUID) (*sqlc.GetInboxMessageDetailByIDRow, error)
	InboxMessageExists(ctx context.Context, inboxID uuid.UUID, messageID string) (bool, error)
	InsertInboxMessage(ctx context.Context, arg sqlc.InsertInboxMessageRecordParams) (sqlc.InboxMessage, error)
	FindTaskByThreadID(ctx context.Context, threadID string) (*sqlc.FindTaskByEmailThreadIDRow, error)
	CreateCommentFromMessage(ctx context.Context, taskID, authorID uuid.UUID, content, emailMessageID string) error
	CreateTaskFromMessageTx(ctx context.Context, msg *sqlc.GetInboxMessageDetailByIDRow, taskType sqlc.TaskType, priority sqlc.TaskPriority, userID uuid.UUID) (*sqlc.InsertTaskFromInboxMessageRow, error)
	MarkMessageConverted(ctx context.Context, messageID, userID uuid.UUID) error
	UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, status sqlc.MessageStatus) (sqlc.InboxMessage, error)

	CreateRule(ctx context.Context, arg sqlc.CreateInboxRuleRecordParams) (sqlc.InboxRule, error)
	ListRulesByInboxID(ctx context.Context, inboxID uuid.UUID) ([]sqlc.InboxRule, error)
	UpdateRule(ctx context.Context, arg sqlc.UpdateInboxRuleByIDParams) (*sqlc.InboxRule, error)
	DeleteRule(ctx context.Context, ruleID uuid.UUID) (int64, error)

	GetCommentReplyContext(ctx context.Context, commentID uuid.UUID) (*sqlc.GetCommentEmailReplyContextRow, error)
	MarkCommentSentAsEmail(ctx context.Context, commentID uuid.UUID, messageID string, recipients []string) error
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL-backed inbox repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{
		db:      db,
		queries: sqlc.New(db.Pool()),
	}
}

func (r *pgRepository) GetProjectByID(ctx context.Context, projectID uuid.UUID) (*sqlc.GetProjectByIDForInboxRow, error) {
	row, err := r.queries.GetProjectByIDForInbox(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetInboxByProjectID(ctx context.Context, projectID uuid.UUID) (*sqlc.GetProjectInboxByProjectIDRow, error) {
	row, err := r.queries.GetProjectInboxByProjectID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) FindInboxByEmail(ctx context.Context, email string) (*sqlc.GetProjectInboxByEmailAddressRow, error) {
	row, err := r.queries.GetProjectInboxByEmailAddress(ctx, pgtype.Text{String: email, Valid: email != ""})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) CreateInbox(ctx context.Context, arg sqlc.CreateProjectInboxRecordParams) (sqlc.ProjectInbox, error) {
	row, err := r.queries.CreateProjectInboxRecord(ctx, arg)
	if err != nil {
		return sqlc.ProjectInbox{}, apperror.Map(err)
	}
	return row, nil
}

func (r *pgRepository) UpdateInbox(ctx context.Context, arg sqlc.UpdateProjectInboxRecordParams) (sqlc.ProjectInbox, error) {
	row, err := r.queries.UpdateProjectInboxRecord(ctx, arg)
	if err != nil {
		return sqlc.ProjectInbox{}, apperror.Map(err)
	}
	return row, nil
}

func (r *pgRepository) UpdateInboxEmailAddress(ctx context.Context, inboxID uuid.UUID, email string) error {
	return r.queries.UpdateInboxEmailAddressByID(ctx, sqlc.UpdateInboxEmailAddressByIDParams{
		ID:           inboxID,
		EmailAddress: pgtype.Text{String: email, Valid: email != ""},
	})
}

func (r *pgRepository) GetEmailAccountByInboxID(ctx context.Context, inboxID uuid.UUID) (*sqlc.EmailAccount, error) {
	row, err := r.queries.GetEmailAccountByInboxID(ctx, inboxID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetEmailAccountByID(ctx context.Context, accountID uuid.UUID) (*sqlc.EmailAccount, error) {
	row, err := r.queries.GetEmailAccountByID(ctx, accountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) UpsertEmailAccount(ctx context.Context, arg sqlc.UpsertEmailAccountForInboxParams) (sqlc.EmailAccount, error) {
	row, err := r.queries.UpsertEmailAccountForInbox(ctx, arg)
	if err != nil {
		return sqlc.EmailAccount{}, apperror.Map(err)
	}
	return row, nil
}

func (r *pgRepository) ListSyncEnabledAccounts(ctx context.Context) ([]sqlc.ListSyncEnabledEmailAccountsRow, error) {
	return r.queries.ListSyncEnabledEmailAccounts(ctx)
}

func (r *pgRepository) GetAccountSyncDetailsByProjectID(ctx context.Context, projectID uuid.UUID) (*sqlc.GetEmailAccountSyncDetailsByProjectIDRow, error) {
	row, err := r.queries.GetEmailAccountSyncDetailsByProjectID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) UpdateEmailAccountSyncState(
	ctx context.Context, accountID uuid.UUID, lastSync time.Time, lastErr *string, uidNext *int32, uidValidity *string,
) error {
	var errTxt, validTxt pgtype.Text
	if lastErr != nil {
		errTxt = pgtype.Text{String: *lastErr, Valid: true}
	}
	if uidValidity != nil {
		validTxt = pgtype.Text{String: *uidValidity, Valid: true}
	}
	var uidN pgtype.Int4
	if uidNext != nil {
		uidN = pgtype.Int4{Int32: *uidNext, Valid: true}
	}
	return r.queries.UpdateEmailAccountSyncState(ctx, sqlc.UpdateEmailAccountSyncStateParams{
		ID:            accountID,
		LastSyncAt:    pgtype.Timestamp{Time: lastSync.UTC(), Valid: true},
		LastSyncError: errTxt,
		UidNext:       uidN,
		UidValidity:   validTxt,
	})
}

func (r *pgRepository) CreateSyncLog(ctx context.Context, arg sqlc.CreateEmailSyncLogParams) error {
	_, err := r.queries.CreateEmailSyncLog(ctx, arg)
	return err
}

func (r *pgRepository) ListMessages(ctx context.Context, arg sqlc.ListInboxMessagesFilteredParams) ([]sqlc.ListInboxMessagesFilteredRow, error) {
	return r.queries.ListInboxMessagesFiltered(ctx, arg)
}

func (r *pgRepository) ListMessageAttachmentsBatch(ctx context.Context, messageIDs []uuid.UUID) ([]sqlc.ListMessageAttachmentsBatchRow, error) {
	if len(messageIDs) == 0 {
		return nil, nil
	}
	return r.queries.ListMessageAttachmentsBatch(ctx, messageIDs)
}

func (r *pgRepository) GetMessageDetailByID(ctx context.Context, messageID uuid.UUID) (*sqlc.GetInboxMessageDetailByIDRow, error) {
	row, err := r.queries.GetInboxMessageDetailByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) InboxMessageExists(ctx context.Context, inboxID uuid.UUID, messageID string) (bool, error) {
	_, err := r.queries.FindInboxMessageByMessageID(ctx, sqlc.FindInboxMessageByMessageIDParams{
		ProjectInboxID: inboxID,
		MessageID:      messageID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *pgRepository) InsertInboxMessage(ctx context.Context, arg sqlc.InsertInboxMessageRecordParams) (sqlc.InboxMessage, error) {
	return r.queries.InsertInboxMessageRecord(ctx, arg)
}

func (r *pgRepository) FindTaskByThreadID(ctx context.Context, threadID string) (*sqlc.FindTaskByEmailThreadIDRow, error) {
	row, err := r.queries.FindTaskByEmailThreadID(ctx, pgtype.Text{String: threadID, Valid: threadID != ""})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) CreateCommentFromMessage(ctx context.Context, taskID, authorID uuid.UUID, content, emailMessageID string) error {
	_, err := r.queries.InsertCommentFromEmailMessage(ctx, sqlc.InsertCommentFromEmailMessageParams{
		ID:             uuid.New(),
		TaskID:         taskID,
		AuthorID:       authorID,
		Content:        content,
		EmailMessageID: pgtype.Text{String: emailMessageID, Valid: emailMessageID != ""},
	})
	return err
}

func (r *pgRepository) CreateTaskFromMessageTx(
	ctx context.Context,
	msg *sqlc.GetInboxMessageDetailByIDRow,
	taskType sqlc.TaskType,
	priority sqlc.TaskPriority,
	userID uuid.UUID,
) (*sqlc.InsertTaskFromInboxMessageRow, error) {
	var created sqlc.InsertTaskFromInboxMessageRow
	err := r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		lastNum, err := q.GetLastTaskNumberByProjectForInbox(ctx, msg.ProjectID)
		if err != nil {
			return err
		}
		nextNum := lastNum + 1
		slug := fmt.Sprintf("%s-%d", msg.ProjectSlug, nextNum)

		desc := msg.BodyText
		if !desc.Valid || desc.String == "" {
			desc = msg.BodyHtml
		}
		taskRow, insErr := q.InsertTaskFromInboxMessage(ctx, sqlc.InsertTaskFromInboxMessageParams{
			ID:             uuid.New(),
			ProjectID:      msg.ProjectID,
			Title:          msg.Subject,
			Description:    desc,
			Type:           taskType,
			Priority:       priority,
			StatusID:       msg.DefaultStatusID,
			EmailThreadID:  pgtype.Text{String: msg.ThreadID, Valid: msg.ThreadID != ""},
			InboxMessageID: pgtype.UUID{Bytes: msg.ID, Valid: true},
			TaskNumber:     nextNum,
			Slug:           slug,
			CreatedByID:    pgtype.UUID{Bytes: userID, Valid: userID != uuid.Nil},
		})
		if insErr != nil {
			return insErr
		}
		created = taskRow

		if msg.DefaultAssigneeID.Valid {
			if aErr := q.InsertTaskAssignee(ctx, sqlc.InsertTaskAssigneeParams{
				TaskID: taskRow.ID,
				UserID: uuid.UUID(msg.DefaultAssigneeID.Bytes),
			}); aErr != nil {
				return aErr
			}
		}
		if rErr := q.InsertTaskReporter(ctx, sqlc.InsertTaskReporterParams{
			TaskID: taskRow.ID,
			UserID: userID,
		}); rErr != nil {
			return rErr
		}
		if mErr := q.MarkInboxMessageConverted(ctx, sqlc.MarkInboxMessageConvertedParams{
			ID:          msg.ID,
			ConvertedBy: pgtype.UUID{Bytes: userID, Valid: userID != uuid.Nil},
		}); mErr != nil {
			return mErr
		}
		if cErr := q.CopyMessageAttachmentsToTask(ctx, sqlc.CopyMessageAttachmentsToTaskParams{
			MessageID: msg.ID,
			TaskID:    taskRow.ID,
		}); cErr != nil {
			return cErr
		}
		return task_ranks.SeedForTasksBatchWithQueries(ctx, q, []uuid.UUID{taskRow.ID}, msg.ProjectID, msg.WorkspaceID, msg.OrganizationID)
	})
	if err != nil {
		return nil, apperror.Map(err)
	}
	return &created, nil
}

func (r *pgRepository) MarkMessageConverted(ctx context.Context, messageID, userID uuid.UUID) error {
	return r.queries.MarkInboxMessageConverted(ctx, sqlc.MarkInboxMessageConvertedParams{
		ID:          messageID,
		ConvertedBy: pgtype.UUID{Bytes: userID, Valid: userID != uuid.Nil},
	})
}

func (r *pgRepository) UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, status sqlc.MessageStatus) (sqlc.InboxMessage, error) {
	row, err := r.queries.UpdateInboxMessageStatusByID(ctx, sqlc.UpdateInboxMessageStatusByIDParams{
		ID:     messageID,
		Status: status,
	})
	if err != nil {
		return sqlc.InboxMessage{}, apperror.Map(err)
	}
	return row, nil
}

func (r *pgRepository) CreateRule(ctx context.Context, arg sqlc.CreateInboxRuleRecordParams) (sqlc.InboxRule, error) {
	row, err := r.queries.CreateInboxRuleRecord(ctx, arg)
	if err != nil {
		return sqlc.InboxRule{}, apperror.Map(err)
	}
	return row, nil
}

func (r *pgRepository) ListRulesByInboxID(ctx context.Context, inboxID uuid.UUID) ([]sqlc.InboxRule, error) {
	return r.queries.ListInboxRulesByInboxID(ctx, inboxID)
}

func (r *pgRepository) UpdateRule(ctx context.Context, arg sqlc.UpdateInboxRuleByIDParams) (*sqlc.InboxRule, error) {
	row, err := r.queries.UpdateInboxRuleByID(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, apperror.Map(err)
	}
	return &row, nil
}

func (r *pgRepository) DeleteRule(ctx context.Context, ruleID uuid.UUID) (int64, error) {
	return r.queries.DeleteInboxRuleByID(ctx, ruleID)
}

func (r *pgRepository) GetCommentReplyContext(ctx context.Context, commentID uuid.UUID) (*sqlc.GetCommentEmailReplyContextRow, error) {
	row, err := r.queries.GetCommentEmailReplyContext(ctx, commentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) MarkCommentSentAsEmail(ctx context.Context, commentID uuid.UUID, messageID string, recipients []string) error {
	return r.queries.MarkCommentSentAsEmail(ctx, sqlc.MarkCommentSentAsEmailParams{
		ID:              commentID,
		EmailMessageID:  pgtype.Text{String: messageID, Valid: true},
		EmailRecipients: recipients,
	})
}

func encodeJSONMap(m map[string]any) ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

func decodeJSONMap(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}
