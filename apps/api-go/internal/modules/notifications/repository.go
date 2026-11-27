package notifications

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines database persistence for notifications.
type Repository interface {
	Create(ctx context.Context, arg sqlc.CreateNotificationRecordParams) (sqlc.Notification, error)
	ListUser(ctx context.Context, arg sqlc.ListUserNotificationsParams) ([]sqlc.ListUserNotificationsRow, error)
	CountUser(ctx context.Context, arg sqlc.CountUserNotificationsParams) (int64, error)
	ListUserOrgDetailed(ctx context.Context, arg sqlc.ListNotificationsByUserOrgDetailedParams) ([]sqlc.ListNotificationsByUserOrgDetailedRow, error)
	CountUserOrgDetailed(ctx context.Context, arg sqlc.CountNotificationsByUserOrgDetailedParams) (int64, error)
	GroupByTypeForUserOrg(ctx context.Context, arg sqlc.GroupNotificationsByTypeForUserOrgParams) ([]sqlc.GroupNotificationsByTypeForUserOrgRow, error)
	GroupByPriorityForUserOrg(ctx context.Context, arg sqlc.GroupNotificationsByPriorityForUserOrgParams) ([]sqlc.GroupNotificationsByPriorityForUserOrgRow, error)
	GetByIDForUser(ctx context.Context, id, userID uuid.UUID) (*sqlc.GetNotificationByIDForUserRow, error)
	MarkAsRead(ctx context.Context, id, userID uuid.UUID) (int64, error)
	MarkAllAsRead(ctx context.Context, arg sqlc.MarkAllNotificationsAsReadParams) (int64, error)
	CountUnread(ctx context.Context, arg sqlc.CountUnreadNotificationsParams) (int64, error)
	CountUnreadByOrg(ctx context.Context, userID uuid.UUID) ([]sqlc.CountUnreadNotificationsByOrganizationRow, error)
	GroupByTypeForUser(ctx context.Context, arg sqlc.GroupNotificationsByTypeForUserParams) ([]sqlc.GroupNotificationsByTypeForUserRow, error)
	DeleteForUser(ctx context.Context, id, userID uuid.UUID) (int64, error)
	DeleteMultipleForUser(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) (int64, error)

	// Entity hydration queries
	GetEntityTask(ctx context.Context, id uuid.UUID) (*sqlc.GetNotificationEntityTaskRow, error)
	GetEntityTaskComment(ctx context.Context, id uuid.UUID) (*sqlc.GetNotificationEntityTaskCommentRow, error)
	GetEntityProject(ctx context.Context, id uuid.UUID) (*sqlc.GetNotificationEntityProjectRow, error)
	GetEntityWorkspace(ctx context.Context, id uuid.UUID) (*sqlc.GetNotificationEntityWorkspaceRow, error)
	GetEntityOrganization(ctx context.Context, id uuid.UUID) (*sqlc.GetNotificationEntityOrganizationRow, error)
	GetEntitySprint(ctx context.Context, id uuid.UUID) (*sqlc.GetNotificationEntitySprintRow, error)
	GetEntityInvitation(ctx context.Context, id uuid.UUID) (*sqlc.GetNotificationEntityInvitationRow, error)
	GetEntityTaskAttachment(ctx context.Context, id uuid.UUID) (*sqlc.GetNotificationEntityTaskAttachmentRow, error)

	// Task & mention helpers
	GetTaskContext(ctx context.Context, taskID uuid.UUID) (*sqlc.GetTaskNotificationContextRow, error)
	ListTaskParticipants(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListTaskNotificationParticipantsRow, error)
	FindMentionedUsers(ctx context.Context, usernames []string, userIDs []uuid.UUID) ([]sqlc.FindMentionedUsersRow, error)
	GetUserBrief(ctx context.Context, userID uuid.UUID) (*sqlc.GetUserBriefForNotificationRow, error)
}

type pgRepository struct {
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL-backed notifications repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{queries: sqlc.New(db.Pool())}
}

func (r *pgRepository) Create(ctx context.Context, arg sqlc.CreateNotificationRecordParams) (sqlc.Notification, error) {
	return r.queries.CreateNotificationRecord(ctx, arg)
}

func (r *pgRepository) ListUser(ctx context.Context, arg sqlc.ListUserNotificationsParams) ([]sqlc.ListUserNotificationsRow, error) {
	return r.queries.ListUserNotifications(ctx, arg)
}

func (r *pgRepository) CountUser(ctx context.Context, arg sqlc.CountUserNotificationsParams) (int64, error) {
	return r.queries.CountUserNotifications(ctx, arg)
}

func (r *pgRepository) ListUserOrgDetailed(ctx context.Context, arg sqlc.ListNotificationsByUserOrgDetailedParams) ([]sqlc.ListNotificationsByUserOrgDetailedRow, error) {
	return r.queries.ListNotificationsByUserOrgDetailed(ctx, arg)
}

func (r *pgRepository) CountUserOrgDetailed(ctx context.Context, arg sqlc.CountNotificationsByUserOrgDetailedParams) (int64, error) {
	return r.queries.CountNotificationsByUserOrgDetailed(ctx, arg)
}

func (r *pgRepository) GroupByTypeForUserOrg(ctx context.Context, arg sqlc.GroupNotificationsByTypeForUserOrgParams) ([]sqlc.GroupNotificationsByTypeForUserOrgRow, error) {
	return r.queries.GroupNotificationsByTypeForUserOrg(ctx, arg)
}

func (r *pgRepository) GroupByPriorityForUserOrg(ctx context.Context, arg sqlc.GroupNotificationsByPriorityForUserOrgParams) ([]sqlc.GroupNotificationsByPriorityForUserOrgRow, error) {
	return r.queries.GroupNotificationsByPriorityForUserOrg(ctx, arg)
}

func (r *pgRepository) GetByIDForUser(ctx context.Context, id, userID uuid.UUID) (*sqlc.GetNotificationByIDForUserRow, error) {
	row, err := r.queries.GetNotificationByIDForUser(ctx, sqlc.GetNotificationByIDForUserParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) MarkAsRead(ctx context.Context, id, userID uuid.UUID) (int64, error) {
	return r.queries.MarkNotificationAsRead(ctx, sqlc.MarkNotificationAsReadParams{
		ID:     id,
		UserID: userID,
	})
}

func (r *pgRepository) MarkAllAsRead(ctx context.Context, arg sqlc.MarkAllNotificationsAsReadParams) (int64, error) {
	return r.queries.MarkAllNotificationsAsRead(ctx, arg)
}

func (r *pgRepository) CountUnread(ctx context.Context, arg sqlc.CountUnreadNotificationsParams) (int64, error) {
	return r.queries.CountUnreadNotifications(ctx, arg)
}

func (r *pgRepository) CountUnreadByOrg(ctx context.Context, userID uuid.UUID) ([]sqlc.CountUnreadNotificationsByOrganizationRow, error) {
	return r.queries.CountUnreadNotificationsByOrganization(ctx, userID)
}

func (r *pgRepository) GroupByTypeForUser(ctx context.Context, arg sqlc.GroupNotificationsByTypeForUserParams) ([]sqlc.GroupNotificationsByTypeForUserRow, error) {
	return r.queries.GroupNotificationsByTypeForUser(ctx, arg)
}

func (r *pgRepository) DeleteForUser(ctx context.Context, id, userID uuid.UUID) (int64, error) {
	return r.queries.DeleteNotificationForUser(ctx, sqlc.DeleteNotificationForUserParams{
		ID:     id,
		UserID: userID,
	})
}

func (r *pgRepository) DeleteMultipleForUser(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) (int64, error) {
	return r.queries.DeleteMultipleNotificationsForUser(ctx, sqlc.DeleteMultipleNotificationsForUserParams{
		UserID: userID,
		Ids:    ids,
	})
}

func (r *pgRepository) GetEntityTask(ctx context.Context, id uuid.UUID) (*sqlc.GetNotificationEntityTaskRow, error) {
	row, err := r.queries.GetNotificationEntityTask(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetEntityTaskComment(ctx context.Context, id uuid.UUID) (*sqlc.GetNotificationEntityTaskCommentRow, error) {
	row, err := r.queries.GetNotificationEntityTaskComment(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetEntityProject(ctx context.Context, id uuid.UUID) (*sqlc.GetNotificationEntityProjectRow, error) {
	row, err := r.queries.GetNotificationEntityProject(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetEntityWorkspace(ctx context.Context, id uuid.UUID) (*sqlc.GetNotificationEntityWorkspaceRow, error) {
	row, err := r.queries.GetNotificationEntityWorkspace(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetEntityOrganization(ctx context.Context, id uuid.UUID) (*sqlc.GetNotificationEntityOrganizationRow, error) {
	row, err := r.queries.GetNotificationEntityOrganization(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetEntitySprint(ctx context.Context, id uuid.UUID) (*sqlc.GetNotificationEntitySprintRow, error) {
	row, err := r.queries.GetNotificationEntitySprint(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetEntityInvitation(ctx context.Context, id uuid.UUID) (*sqlc.GetNotificationEntityInvitationRow, error) {
	row, err := r.queries.GetNotificationEntityInvitation(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetEntityTaskAttachment(ctx context.Context, id uuid.UUID) (*sqlc.GetNotificationEntityTaskAttachmentRow, error) {
	row, err := r.queries.GetNotificationEntityTaskAttachment(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetTaskContext(ctx context.Context, taskID uuid.UUID) (*sqlc.GetTaskNotificationContextRow, error) {
	row, err := r.queries.GetTaskNotificationContext(ctx, taskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListTaskParticipants(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListTaskNotificationParticipantsRow, error) {
	return r.queries.ListTaskNotificationParticipants(ctx, taskID)
}

func (r *pgRepository) FindMentionedUsers(ctx context.Context, usernames []string, userIDs []uuid.UUID) ([]sqlc.FindMentionedUsersRow, error) {
	return r.queries.FindMentionedUsers(ctx, sqlc.FindMentionedUsersParams{
		Usernames: usernames,
		UserIds:   userIDs,
	})
}

func (r *pgRepository) GetUserBrief(ctx context.Context, userID uuid.UUID) (*sqlc.GetUserBriefForNotificationRow, error) {
	row, err := r.queries.GetUserBriefForNotification(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}
