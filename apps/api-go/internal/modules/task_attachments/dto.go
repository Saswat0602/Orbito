package task_attachments

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// CreatorBrief represents uploader metadata attached to an attachment record.
type CreatorBrief struct {
	ID        string  `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Email     string  `json:"email"`
	Username  *string `json:"username,omitempty"`
	Avatar    *string `json:"avatar"`
}

// StatusBrief represents task status details on attachment detail responses.
type StatusBrief struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Category string `json:"category"`
}

// OrgBrief represents minimal organization metadata.
type OrgBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// WorkspaceBrief represents minimal workspace metadata.
type WorkspaceBrief struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Organization *OrgBrief `json:"organization,omitempty"`
}

// ProjectBrief represents minimal project metadata.
type ProjectBrief struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Slug      string          `json:"slug"`
	Workspace *WorkspaceBrief `json:"workspace,omitempty"`
}

// TaskBrief represents task metadata embedded in attachment responses.
type TaskBrief struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Slug        string        `json:"slug"`
	Description *string       `json:"description,omitempty"`
	TaskNumber  *int32        `json:"taskNumber,omitempty"`
	Priority    *string       `json:"priority,omitempty"`
	Type        *string       `json:"type,omitempty"`
	Status      *StatusBrief  `json:"status,omitempty"`
	Project     *ProjectBrief `json:"project,omitempty"`
}

// AttachmentResponse represents a task attachment returned across upload, list, and detail endpoints.
type AttachmentResponse struct {
	ID            string        `json:"id"`
	FileName      string        `json:"fileName"`
	FilePath      *string       `json:"filePath"`
	FileSize      int32         `json:"fileSize"`
	MimeType      string        `json:"mimeType"`
	URL           *string       `json:"url"`
	StorageKey    *string       `json:"storageKey"`
	TaskID        string        `json:"taskId"`
	CreatedBy     *string       `json:"createdBy"`
	UpdatedBy     *string       `json:"updatedBy"`
	CreatedAt     string        `json:"createdAt"`
	UpdatedAt     string        `json:"updatedAt"`
	ViewURL       *string       `json:"viewUrl,omitempty"`
	Task          *TaskBrief    `json:"task,omitempty"`
	CreatedByUser *CreatorBrief `json:"createdByUser,omitempty"`
}

// CategoryStat represents count and cumulative bytes for a file category.
type CategoryStat struct {
	Count     int64 `json:"count"`
	TotalSize int64 `json:"totalSize"`
}

// StatsResponse matches GET /task-attachments/stats output.
type StatsResponse struct {
	TotalAttachments int64                   `json:"totalAttachments"`
	TotalSizeBytes   int64                   `json:"totalSizeBytes"`
	TotalSizeMB      float64                 `json:"totalSizeMB"`
	FileCategories   map[string]CategoryStat `json:"fileCategories"`
	RecentUploads    int64                   `json:"recentUploads"`
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

func resolveViewURL(urlVal, keyVal *string) *string {
	if urlVal != nil && *urlVal != "" {
		return urlVal
	}
	if keyVal != nil && *keyVal != "" {
		u := "/uploads/" + *keyVal
		return &u
	}
	return nil
}

func mapDetailRow(r *sqlc.GetTaskAttachmentDetailByIDRow, includeDeep bool) AttachmentResponse {
	urlPtr := textPtr(r.Url)
	keyPtr := textPtr(r.StorageKey)
	resp := AttachmentResponse{
		ID:         r.ID.String(),
		FileName:   r.FileName,
		FilePath:   textPtr(r.FilePath),
		FileSize:   r.FileSize,
		MimeType:   r.MimeType,
		URL:        urlPtr,
		StorageKey: keyPtr,
		TaskID:     r.TaskID.String(),
		CreatedBy:  uuidPtr(r.CreatedByID),
		UpdatedBy:  uuidPtr(r.UpdatedByID),
		CreatedAt:  tsStr(r.CreatedAt),
		UpdatedAt:  tsStr(r.UpdatedAt),
		ViewURL:    resolveViewURL(urlPtr, keyPtr),
		Task: &TaskBrief{
			ID:    r.TaskID.String(),
			Title: r.TaskTitle,
			Slug:  r.TaskSlug,
			Project: &ProjectBrief{
				ID:   r.ProjectID.String(),
				Name: r.ProjectName,
				Slug: r.ProjectSlug,
			},
		},
	}
	if includeDeep {
		tn := r.TaskNumber
		pr := string(r.TaskPriority)
		tp := string(r.TaskType)
		resp.Task.Description = textPtr(r.TaskDescription)
		resp.Task.TaskNumber = &tn
		resp.Task.Priority = &pr
		resp.Task.Type = &tp
		resp.Task.Status = &StatusBrief{
			ID:       r.StatusID.String(),
			Name:     r.StatusName,
			Color:    r.StatusColor,
			Category: string(r.StatusCategory),
		}
		resp.Task.Project.Workspace = &WorkspaceBrief{
			ID:   r.WorkspaceID.String(),
			Name: r.WorkspaceName,
			Slug: r.WorkspaceSlug,
			Organization: &OrgBrief{
				ID:   r.OrgID.String(),
				Name: r.OrgName,
				Slug: r.OrgSlug,
			},
		}
		if r.CreatorID.Valid {
			resp.CreatedByUser = &CreatorBrief{
				ID:        uuid.UUID(r.CreatorID.Bytes).String(),
				FirstName: r.CreatorFirstName.String,
				LastName:  r.CreatorLastName.String,
				Email:     r.CreatorEmail.String,
				Username:  textPtr(r.CreatorUsername),
				Avatar:    textPtr(r.CreatorAvatar),
			}
		}
	}
	return resp
}

func mapListRow(r sqlc.ListTaskAttachmentDetailsByTaskIDRow, includeTask bool) AttachmentResponse {
	urlPtr := textPtr(r.Url)
	keyPtr := textPtr(r.StorageKey)
	resp := AttachmentResponse{
		ID:         r.ID.String(),
		FileName:   r.FileName,
		FilePath:   textPtr(r.FilePath),
		FileSize:   r.FileSize,
		MimeType:   r.MimeType,
		URL:        urlPtr,
		StorageKey: keyPtr,
		TaskID:     r.TaskID.String(),
		CreatedBy:  uuidPtr(r.CreatedByID),
		UpdatedBy:  uuidPtr(r.UpdatedByID),
		CreatedAt:  tsStr(r.CreatedAt),
		UpdatedAt:  tsStr(r.UpdatedAt),
		ViewURL:    resolveViewURL(urlPtr, keyPtr),
	}
	if includeTask {
		resp.Task = &TaskBrief{
			ID:    r.TaskID.String(),
			Title: r.TaskTitle,
			Slug:  r.TaskSlug,
			Project: &ProjectBrief{
				ID:   r.ProjectID.String(),
				Name: r.ProjectName,
				Slug: r.ProjectSlug,
			},
		}
	}
	if r.CreatorID.Valid {
		resp.CreatedByUser = &CreatorBrief{
			ID:        uuid.UUID(r.CreatorID.Bytes).String(),
			FirstName: r.CreatorFirstName.String,
			LastName:  r.CreatorLastName.String,
			Email:     r.CreatorEmail.String,
			Avatar:    textPtr(r.CreatorAvatar),
		}
	}
	return resp
}

