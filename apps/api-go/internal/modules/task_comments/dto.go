package task_comments

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// CreateRequest represents CreateTaskCommentDto.
type CreateRequest struct {
	Content         string  `json:"content" validate:"required,min=1"`
	TaskID          string  `json:"taskId" validate:"required"`
	ParentCommentID *string `json:"parentCommentId" validate:"omitempty,uuid"`
}

// UpdateRequest represents UpdateTaskCommentDto.
type UpdateRequest struct {
	Content string `json:"content" validate:"required,min=1"`
}

// AuthorBrief represents author details embedded in comments.
type AuthorBrief struct {
	ID        string  `json:"id"`
	Email     *string `json:"email,omitempty"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Avatar    *string `json:"avatar"`
}

// ProjectBrief represents task.project details for comment findOne.
type ProjectBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// TaskBrief represents task metadata embedded in comments.
type TaskBrief struct {
	ID                string        `json:"id"`
	Title             string        `json:"title"`
	Slug              string        `json:"slug"`
	AllowEmailReplies *bool         `json:"allowEmailReplies,omitempty"`
	Project           *ProjectBrief `json:"project,omitempty"`
}

// ParentCommentBrief represents parent comment summary.
type ParentCommentBrief struct {
	ID        string      `json:"id"`
	Content   string      `json:"content"`
	Author    AuthorBrief `json:"author"`
	CreatedAt *string     `json:"createdAt,omitempty"`
}

// ReplyCount mirrors Prisma `_count: { replies }`.
type ReplyCount struct {
	Replies int64 `json:"replies"`
}

// ReplyResponse represents a reply comment row.
type ReplyResponse struct {
	ID              string          `json:"id"`
	Content         string          `json:"content"`
	TaskID          string          `json:"taskId"`
	AuthorID        string          `json:"authorId"`
	ParentCommentID *string         `json:"parentCommentId"`
	CreatedBy       *string         `json:"createdBy"`
	UpdatedBy       *string         `json:"updatedBy"`
	CreatedAt       string          `json:"createdAt"`
	UpdatedAt       string          `json:"updatedAt"`
	Author          AuthorBrief     `json:"author"`
	Replies         []ReplyResponse `json:"replies,omitempty"`
	Count           *ReplyCount     `json:"_count,omitempty"`
}

// CommentResponse represents a full task comment record.
type CommentResponse struct {
	ID                  string              `json:"id"`
	Content             string              `json:"content"`
	TaskID              string              `json:"taskId"`
	AuthorID            string              `json:"authorId"`
	ParentCommentID     *string             `json:"parentCommentId"`
	CreatedBy           *string             `json:"createdBy"`
	UpdatedBy           *string             `json:"updatedBy"`
	CreatedAt           string              `json:"createdAt"`
	UpdatedAt           string              `json:"updatedAt"`
	EmailMessageID      *string             `json:"emailMessageId"`
	SentAsEmail         bool                `json:"sentAsEmail"`
	EmailRecipients     []string            `json:"emailRecipients"`
	EmailRecipientNames *string             `json:"emailRecipientNames"`
	EmailSentAt         *string             `json:"emailSentAt"`
	Author              AuthorBrief         `json:"author"`
	Task                *TaskBrief          `json:"task,omitempty"`
	ParentComment       *ParentCommentBrief `json:"parentComment,omitempty"`
	Replies             []ReplyResponse     `json:"replies,omitempty"`
	Count               *ReplyCount         `json:"_count,omitempty"`
}

// PaginatedCommentsResponse matches findAll output.
type PaginatedCommentsResponse struct {
	Data       []CommentResponse `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"totalPages"`
	HasMore    bool              `json:"hasMore"`
}

// MiddlePaginatedCommentsResponse matches findWithMiddlePagination output.
type MiddlePaginatedCommentsResponse struct {
	Data        []CommentResponse `json:"data"`
	Total       int64             `json:"total"`
	Page        int               `json:"page"`
	Limit       int               `json:"limit"`
	TotalPages  int               `json:"totalPages"`
	HasMore     bool              `json:"hasMore"`
	LoadedCount int               `json:"loadedCount"`
}

func uuidPtr(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := uuid.UUID(u.Bytes).String()
	return &s
}

func textPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}

func tsStr(ts pgtype.Timestamp) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.UTC().Format("2006-01-02T15:04:05.000Z")
}

func tsPtr(ts pgtype.Timestamp) *string {
	if !ts.Valid {
		return nil
	}
	s := ts.Time.UTC().Format("2006-01-02T15:04:05.000Z")
	return &s
}

func nonNilSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func mapReplyRow(r sqlc.ListRepliesByParentIDsRow) ReplyResponse {
	return ReplyResponse{
		ID:              r.ID.String(),
		Content:         r.Content,
		TaskID:          r.TaskID.String(),
		AuthorID:        r.AuthorID.String(),
		ParentCommentID: uuidPtr(r.ParentCommentID),
		CreatedBy:       uuidPtr(r.CreatedByID),
		UpdatedBy:       uuidPtr(r.UpdatedByID),
		CreatedAt:       tsStr(r.CreatedAt),
		UpdatedAt:       tsStr(r.UpdatedAt),
		Author: AuthorBrief{
			ID:        r.AuthorID.String(),
			FirstName: r.AuthorFirstName,
			LastName:  r.AuthorLastName,
			Avatar:    textPtr(r.AuthorAvatar),
		},
		Count: &ReplyCount{Replies: r.ReplyCount},
	}
}

func mapDetailRow(r *sqlc.GetTaskCommentDetailByIDRow, replies []ReplyResponse, includeProject, includeEmail bool) CommentResponse {
	var emailPtr *string
	if includeEmail {
		e := r.AuthorEmail
		emailPtr = &e
	}
	allowEmail := r.TaskAllowEmailReplies
	resp := CommentResponse{
		ID:                  r.ID.String(),
		Content:             r.Content,
		TaskID:              r.TaskID.String(),
		AuthorID:            r.AuthorID.String(),
		ParentCommentID:     uuidPtr(r.ParentCommentID),
		CreatedBy:           uuidPtr(r.CreatedByID),
		UpdatedBy:           uuidPtr(r.UpdatedByID),
		CreatedAt:           tsStr(r.CreatedAt),
		UpdatedAt:           tsStr(r.UpdatedAt),
		EmailMessageID:      textPtr(r.EmailMessageID),
		SentAsEmail:         r.SentAsEmail,
		EmailRecipients:     nonNilSlice(r.EmailRecipients),
		EmailRecipientNames: textPtr(r.EmailRecipientNames),
		EmailSentAt:         tsPtr(r.EmailSentAt),
		Author: AuthorBrief{
			ID:        r.AuthorID.String(),
			Email:     emailPtr,
			FirstName: r.AuthorFirstName,
			LastName:  r.AuthorLastName,
			Avatar:    textPtr(r.AuthorAvatar),
		},
		Task: &TaskBrief{
			ID:                r.TaskID.String(),
			Title:             r.TaskTitle,
			Slug:              r.TaskSlug,
			AllowEmailReplies: &allowEmail,
		},
		Replies: replies,
		Count:   &ReplyCount{Replies: r.ReplyCount},
	}
	if includeProject {
		resp.Task.AllowEmailReplies = nil
		resp.Task.Project = &ProjectBrief{
			ID:   r.ProjectID.String(),
			Name: r.ProjectName,
			Slug: r.ProjectSlug,
		}
	}
	if r.ParentID.Valid {
		created := tsStr(r.ParentCreatedAt)
		resp.ParentComment = &ParentCommentBrief{
			ID:      uuid.UUID(r.ParentID.Bytes).String(),
			Content: r.ParentContent.String,
			Author: AuthorBrief{
				ID:        uuid.UUID(r.ParentAuthorID.Bytes).String(),
				FirstName: r.ParentAuthorFirstName.String,
				LastName:  r.ParentAuthorLastName.String,
				Avatar:    textPtr(r.ParentAuthorAvatar),
			},
			CreatedAt: &created,
		}
	}
	return resp
}

func mapTopLevelDescRow(r sqlc.ListTopLevelTaskCommentsDescRow, replies []ReplyResponse, includeEmail bool) CommentResponse {
	var emailPtr *string
	if includeEmail {
		e := r.AuthorEmail
		emailPtr = &e
	}
	return CommentResponse{
		ID:                  r.ID.String(),
		Content:             r.Content,
		TaskID:              r.TaskID.String(),
		AuthorID:            r.AuthorID.String(),
		ParentCommentID:     uuidPtr(r.ParentCommentID),
		CreatedBy:           uuidPtr(r.CreatedByID),
		UpdatedBy:           uuidPtr(r.UpdatedByID),
		CreatedAt:           tsStr(r.CreatedAt),
		UpdatedAt:           tsStr(r.UpdatedAt),
		EmailMessageID:      textPtr(r.EmailMessageID),
		SentAsEmail:         r.SentAsEmail,
		EmailRecipients:     nonNilSlice(r.EmailRecipients),
		EmailRecipientNames: textPtr(r.EmailRecipientNames),
		EmailSentAt:         tsPtr(r.EmailSentAt),
		Author: AuthorBrief{
			ID:        r.AuthorID.String(),
			Email:     emailPtr,
			FirstName: r.AuthorFirstName,
			LastName:  r.AuthorLastName,
			Avatar:    textPtr(r.AuthorAvatar),
		},
		Task: &TaskBrief{
			ID:    r.TaskID.String(),
			Title: r.TaskTitle,
			Slug:  r.TaskSlug,
		},
		Replies: replies,
		Count:   &ReplyCount{Replies: r.ReplyCount},
	}
}

