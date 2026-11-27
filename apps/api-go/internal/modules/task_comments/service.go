package task_comments

import (
	"context"
	"math"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/sanitize"
)

// EventEmitter emits comment creation events (optional, wired in Phase 5).
type EventEmitter interface {
	EmitCommentAdded(projectID, taskID string, comment any)
}

// NotificationHooks delivers mention/comment notifications & outbound email replies (optional, wired in Phases 5-6).
type NotificationHooks interface {
	OnCommentCreated(ctx context.Context, commentID, taskID, authorID uuid.UUID, content string)
	OnCommentUpdated(ctx context.Context, commentID, taskID, authorID uuid.UUID, oldContent, newContent string)
}

// Service defines business operations for task comments.
type Service interface {
	Create(ctx context.Context, req CreateRequest, userID uuid.UUID) (*CommentResponse, error)
	FindAll(ctx context.Context, taskIdentifier string, userID uuid.UUID, page, limit int, sort string) (*PaginatedCommentsResponse, error)
	FindWithMiddlePagination(ctx context.Context, taskIdentifier string, userID uuid.UUID, page, limit, oldestCount, newestCount int) (*MiddlePaginatedCommentsResponse, error)
	FindOne(ctx context.Context, id, userID uuid.UUID) (*CommentResponse, error)
	GetReplies(ctx context.Context, commentID, userID uuid.UUID) ([]ReplyResponse, error)
	GetTaskCommentTree(ctx context.Context, taskIdentifier string, userID uuid.UUID) ([]CommentResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest, userID uuid.UUID) (*CommentResponse, error)
	Remove(ctx context.Context, id, userID uuid.UUID) error
}

type service struct {
	repo     Repository
	events   EventEmitter
	notifier NotificationHooks
}

// NewService constructs a task_comments service.
func NewService(repo Repository, events EventEmitter, notifier NotificationHooks) Service {
	return &service{repo: repo, events: events, notifier: notifier}
}

func (s *service) checkAccess(ctx context.Context, userID uuid.UUID, taskIdentifier string) (*sqlc.CheckCommentAccessTaskAndUserRow, error) {
	row, err := s.repo.CheckAccess(ctx, userID, taskIdentifier)
	if err != nil {
		return nil, apperror.Internal("Failed to verify task access")
	}
	if row == nil {
		return nil, apperror.NotFound("Task not found")
	}
	if row.OrgArchived {
		return nil, apperror.Forbidden("Operation not allowed in an archived organization")
	}
	if !row.UserExists {
		return nil, apperror.NotFound("User not found")
	}
	if !row.ProjectRole.Valid && !row.WorkspaceRole.Valid && !row.OrganizationRole.Valid {
		return nil, apperror.Forbidden("You do not have access to this task")
	}
	return row, nil
}

func isAdminRole(role sqlc.NullRole) bool {
	if !role.Valid {
		return false
	}
	return role.Role == sqlc.RoleOWNER || role.Role == sqlc.RoleMANAGER || role.Role == sqlc.RoleSUPERADMIN
}

func safeInt32(n int) int32 {
	if n < 0 {
		return 0
	}
	if n > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(n)
}

func (s *service) attachRepliesToDescRows(ctx context.Context, rows []sqlc.ListTopLevelTaskCommentsDescRow, includeEmail bool) ([]CommentResponse, error) {
	parentIDs := make([]uuid.UUID, len(rows))
	for i, r := range rows {
		parentIDs[i] = r.ID
	}
	replyRows, err := s.repo.ListRepliesByParentIDs(ctx, parentIDs)
	if err != nil {
		return nil, apperror.Internal("Failed to load comment replies")
	}
	replyMap := make(map[uuid.UUID][]ReplyResponse)
	for _, rr := range replyRows {
		if rr.ParentCommentID.Valid {
			pid := uuid.UUID(rr.ParentCommentID.Bytes)
			replyMap[pid] = append(replyMap[pid], mapReplyRow(rr))
		}
	}
	out := make([]CommentResponse, 0, len(rows))
	for _, r := range rows {
		replies := replyMap[r.ID]
		if replies == nil {
			replies = []ReplyResponse{}
		}
		out = append(out, mapTopLevelDescRow(r, replies, includeEmail))
	}
	return out, nil
}

func (s *service) Create(ctx context.Context, req CreateRequest, userID uuid.UUID) (*CommentResponse, error) {
	acc, err := s.checkAccess(ctx, userID, req.TaskID)
	if err != nil {
		return nil, err
	}
	if acc.IsArchived {
		return nil, apperror.Forbidden("Cannot add comments to an archived task")
	}

	var parentUUID *uuid.UUID
	if req.ParentCommentID != nil && *req.ParentCommentID != "" {
		parsed, pErr := uuid.Parse(*req.ParentCommentID)
		if pErr != nil {
			return nil, apperror.BadRequest("parentCommentId must be a UUID")
		}
		parentRow, getErr := s.repo.GetBasicByID(ctx, parsed)
		if getErr != nil {
			return nil, apperror.Internal("Failed to check parent comment")
		}
		if parentRow == nil {
			return nil, apperror.NotFound("Parent comment not found")
		}
		if parentRow.TaskID != acc.TaskID {
			return nil, apperror.BadRequest("Parent comment must belong to the same task")
		}
		parentUUID = &parsed
	}

	cleanContent := sanitize.HTML(req.Content)
	newID, err := s.repo.Create(ctx, uuid.New(), cleanContent, acc.TaskID, userID, parentUUID)
	if err != nil {
		return nil, apperror.Internal("Failed to create comment")
	}

	detail, err := s.repo.GetDetailByID(ctx, newID)
	if err != nil || detail == nil {
		return nil, apperror.Internal("Failed to load created comment")
	}
	resp := mapDetailRow(detail, nil, false, true)

	if s.notifier != nil {
		s.notifier.OnCommentCreated(ctx, newID, acc.TaskID, userID, cleanContent)
	}
	if s.events != nil {
		s.events.EmitCommentAdded(acc.ProjectID.String(), acc.TaskID.String(), resp)
	}
	return &resp, nil
}

func (s *service) FindAll(ctx context.Context, taskIdentifier string, userID uuid.UUID, page, limit int, sort string) (*PaginatedCommentsResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if sort != "asc" {
		sort = "desc"
	}

	acc, err := s.checkAccess(ctx, userID, taskIdentifier)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.CountTopLevel(ctx, acc.TaskID)
	if err != nil {
		return nil, apperror.Internal("Failed to count comments")
	}
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	offset := safeInt32((page - 1) * limit)
	lim32 := safeInt32(limit)

	var descRows []sqlc.ListTopLevelTaskCommentsDescRow
	if sort == "asc" {
		ascRows, ascErr := s.repo.ListTopLevelAsc(ctx, acc.TaskID, lim32, offset)
		if ascErr != nil {
			return nil, apperror.Internal("Failed to list comments")
		}
		descRows = make([]sqlc.ListTopLevelTaskCommentsDescRow, len(ascRows))
		for i, r := range ascRows {
			descRows[i] = sqlc.ListTopLevelTaskCommentsDescRow(r)
		}
	} else {
		var descErr error
		descRows, descErr = s.repo.ListTopLevelDesc(ctx, acc.TaskID, lim32, offset)
		if descErr != nil {
			return nil, apperror.Internal("Failed to list comments")
		}
	}

	data, err := s.attachRepliesToDescRows(ctx, descRows, true)
	if err != nil {
		return nil, err
	}

	return &PaginatedCommentsResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		HasMore:    page < totalPages,
	}, nil
}

func (s *service) FindWithMiddlePagination(ctx context.Context, taskIdentifier string, userID uuid.UUID, page, limit, oldestCount, newestCount int) (*MiddlePaginatedCommentsResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 5
	}
	if oldestCount < 0 {
		oldestCount = 2
	}
	if newestCount < 0 {
		newestCount = 2
	}

	acc, err := s.checkAccess(ctx, userID, taskIdentifier)
	if err != nil {
		return nil, err
	}
	total, err := s.repo.CountTopLevel(ctx, acc.TaskID)
	if err != nil {
		return nil, apperror.Internal("Failed to count comments")
	}

	middleCount := int(total) - oldestCount - newestCount
	if middleCount < 0 {
		middleCount = 0
	}

	var combined []sqlc.ListTopLevelTaskCommentsDescRow
	loadedCount := 0

	if page == 1 {
		if int(total) <= oldestCount+newestCount {
			ascRows, ascErr := s.repo.ListTopLevelAsc(ctx, acc.TaskID, safeInt32(int(total)+1), 0)
			if ascErr != nil {
				return nil, apperror.Internal("Failed to list comments")
			}
			for _, r := range ascRows {
				combined = append(combined, sqlc.ListTopLevelTaskCommentsDescRow(r))
			}
		} else {
			oldest, oErr := s.repo.ListTopLevelAsc(ctx, acc.TaskID, safeInt32(oldestCount), 0)
			if oErr != nil {
				return nil, apperror.Internal("Failed to list oldest comments")
			}
			newest, nErr := s.repo.ListTopLevelDesc(ctx, acc.TaskID, safeInt32(newestCount), 0)
			if nErr != nil {
				return nil, apperror.Internal("Failed to list newest comments")
			}
			for _, r := range oldest {
				combined = append(combined, sqlc.ListTopLevelTaskCommentsDescRow(r))
			}
			for i := len(newest) - 1; i >= 0; i-- {
				combined = append(combined, newest[i])
			}
		}
	} else {
		endIndex := int(total) - newestCount
		skip := oldestCount + (page-2)*limit
		remaining := endIndex - skip
		if remaining < 0 {
			remaining = 0
		}
		take := limit
		if remaining < take {
			take = remaining
		}
		if take > 0 && skip >= 0 {
			midRows, mErr := s.repo.ListTopLevelAsc(ctx, acc.TaskID, safeInt32(take), safeInt32(skip))
			if mErr != nil {
				return nil, apperror.Internal("Failed to list middle comments")
			}
			for _, r := range midRows {
				combined = append(combined, sqlc.ListTopLevelTaskCommentsDescRow(r))
			}
		}
		loadedCount = (page - 1) * limit
		if middleCount < loadedCount {
			loadedCount = middleCount
		}
	}

	data, err := s.attachRepliesToDescRows(ctx, combined, true)
	if err != nil {
		return nil, err
	}
	middlePages := (middleCount + limit - 1) / limit
	totalPages := middlePages + 1

	return &MiddlePaginatedCommentsResponse{
		Data:        data,
		Total:       total,
		Page:        page,
		Limit:       limit,
		TotalPages:  totalPages,
		HasMore:     page < totalPages,
		LoadedCount: loadedCount,
	}, nil
}

func (s *service) FindOne(ctx context.Context, id, userID uuid.UUID) (*CommentResponse, error) {
	detail, err := s.repo.GetDetailByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch comment")
	}
	if detail == nil {
		return nil, apperror.NotFound("Comment not found")
	}
	if _, accErr := s.checkAccess(ctx, userID, detail.TaskID.String()); accErr != nil {
		return nil, accErr
	}
	replyRows, err := s.repo.ListRepliesByParentIDs(ctx, []uuid.UUID{id})
	if err != nil {
		return nil, apperror.Internal("Failed to fetch comment replies")
	}
	replies := make([]ReplyResponse, 0, len(replyRows))
	for _, rr := range replyRows {
		replies = append(replies, mapReplyRow(rr))
	}
	resp := mapDetailRow(detail, replies, true, true)
	return &resp, nil
}

func (s *service) GetReplies(ctx context.Context, commentID, userID uuid.UUID) ([]ReplyResponse, error) {
	parent, err := s.repo.GetBasicByID(ctx, commentID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch parent comment")
	}
	if parent == nil {
		return nil, apperror.NotFound("Comment not found")
	}
	if _, accErr := s.checkAccess(ctx, userID, parent.TaskID.String()); accErr != nil {
		return nil, accErr
	}

	firstLevel, err := s.repo.ListRepliesByParentIDs(ctx, []uuid.UUID{commentID})
	if err != nil {
		return nil, apperror.Internal("Failed to fetch replies")
	}
	ids := make([]uuid.UUID, len(firstLevel))
	for i, r := range firstLevel {
		ids[i] = r.ID
	}
	nestedRows, err := s.repo.ListRepliesByParentIDs(ctx, ids)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch nested replies")
	}
	nestedMap := make(map[uuid.UUID][]ReplyResponse)
	for _, nr := range nestedRows {
		if nr.ParentCommentID.Valid {
			pid := uuid.UUID(nr.ParentCommentID.Bytes)
			nestedMap[pid] = append(nestedMap[pid], mapReplyRow(nr))
		}
	}

	out := make([]ReplyResponse, 0, len(firstLevel))
	for _, r := range firstLevel {
		mapped := mapReplyRow(r)
		if nested := nestedMap[r.ID]; nested != nil {
			mapped.Replies = nested
		} else {
			mapped.Replies = []ReplyResponse{}
		}
		out = append(out, mapped)
	}
	return out, nil
}

func (s *service) GetTaskCommentTree(ctx context.Context, taskIdentifier string, userID uuid.UUID) ([]CommentResponse, error) {
	acc, err := s.checkAccess(ctx, userID, taskIdentifier)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListTopLevelDesc(ctx, acc.TaskID, 1000000, 0)
	if err != nil {
		return nil, apperror.Internal("Failed to list comment tree")
	}
	return s.attachRepliesToDescRows(ctx, rows, false)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest, userID uuid.UUID) (*CommentResponse, error) {
	existing, err := s.repo.GetBasicByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal("Failed to lookup comment")
	}
	if existing == nil {
		return nil, apperror.NotFound("Comment not found")
	}
	acc, err := s.checkAccess(ctx, userID, existing.TaskID.String())
	if err != nil {
		return nil, err
	}
	if acc.IsArchived {
		return nil, apperror.Forbidden("Cannot update comments on an archived task")
	}
	if existing.AuthorID != userID {
		return nil, apperror.Forbidden("You can only edit your own comments")
	}

	cleanContent := sanitize.HTML(req.Content)
	if _, updErr := s.repo.UpdateContent(ctx, id, cleanContent); updErr != nil {
		return nil, apperror.Internal("Failed to update comment")
	}
	detail, err := s.repo.GetDetailByID(ctx, id)
	if err != nil || detail == nil {
		return nil, apperror.Internal("Failed to load updated comment")
	}
	resp := mapDetailRow(detail, nil, false, false)
	if s.notifier != nil {
		s.notifier.OnCommentUpdated(ctx, id, existing.TaskID, userID, existing.Content, cleanContent)
	}
	return &resp, nil
}

func (s *service) Remove(ctx context.Context, id, userID uuid.UUID) error {
	existing, err := s.repo.GetBasicByID(ctx, id)
	if err != nil {
		return apperror.Internal("Failed to lookup comment")
	}
	if existing == nil {
		return apperror.NotFound("Comment not found")
	}
	acc, err := s.checkAccess(ctx, userID, existing.TaskID.String())
	if err != nil {
		return err
	}
	isAdmin := isAdminRole(acc.ProjectRole) || isAdminRole(acc.WorkspaceRole) || isAdminRole(acc.OrganizationRole)
	if existing.AuthorID != userID && !isAdmin {
		return apperror.Forbidden("You can only delete your own comments")
	}
	if _, delErr := s.repo.DeleteByID(ctx, id); delErr != nil {
		return apperror.Internal("Failed to delete comment")
	}
	return nil
}

