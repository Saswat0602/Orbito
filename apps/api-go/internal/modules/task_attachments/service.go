package task_attachments

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/platform/storage"
	"github.com/orbito/api-go/internal/shared/apperror"
)

const maxUploadBytes = 50 * 1024 * 1024

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true, "image/svg+xml": true,
	"application/pdf": true, "application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
	"application/vnd.ms-powerpoint":                                             true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"text/plain": true, "text/csv": true, "text/markdown": true, "application/json": true,
	"application/xml": true, "text/html": true, "text/css": true, "text/javascript": true,
	"application/zip": true, "application/x-rar-compressed": true, "application/x-7z-compressed": true,
	"video/mp4": true, "video/webm": true, "video/ogg": true, "video/mpeg": true,
	"video/quicktime": true, "video/x-msvideo": true, "video/x-matroska": true,
}

var previewableMimeTypes = map[string]bool{
	"image/jpeg": true, "image/jpg": true, "image/png": true, "image/gif": true,
	"image/webp": true, "image/svg+xml": true,
	"video/mp4": true, "video/webm": true, "video/ogg": true, "video/quicktime": true,
	"application/pdf": true, "text/plain": true, "text/html": true, "text/markdown": true,
	"application/json": true, "application/xml": true, "text/css": true, "text/javascript": true,
}

// Service defines business operations for task attachments.
type Service interface {
	Upload(ctx context.Context, taskIdentifier string, userID uuid.UUID, fileName, mimeType string, data []byte) (*AttachmentResponse, error)
	FindAll(ctx context.Context, taskIdentifier string, userID uuid.UUID) ([]AttachmentResponse, error)
	GetTaskAttachments(ctx context.Context, taskIdentifier string, userID uuid.UUID) ([]AttachmentResponse, error)
	FindOne(ctx context.Context, id, userID uuid.UUID) (*AttachmentResponse, error)
	StreamFile(ctx context.Context, id, userID uuid.UUID, w http.ResponseWriter, r *http.Request, isDownload bool) error
	GetStats(ctx context.Context, taskIdentifier string, userID uuid.UUID) (*StatsResponse, error)
	Remove(ctx context.Context, id, userID uuid.UUID) (*AttachmentResponse, error)
}

type service struct {
	repo    Repository
	storage storage.Provider
}

// NewService constructs a task_attachments service.
func NewService(repo Repository, store storage.Provider) Service {
	return &service{repo: repo, storage: store}
}

func (s *service) checkAccess(ctx context.Context, userID uuid.UUID, taskIdentifier string) (*sqlc.CheckAttachmentAccessTaskAndUserRow, error) {
	acc, err := s.repo.CheckAccess(ctx, userID, taskIdentifier)
	if err != nil {
		return nil, apperror.Internal("Failed to check task access")
	}
	if acc == nil {
		return nil, apperror.NotFound("Task not found")
	}
	if !acc.UserExists {
		return nil, apperror.NotFound("User not found")
	}
	if !acc.ProjectRole.Valid && !acc.WorkspaceRole.Valid && !acc.OrganizationRole.Valid {
		return nil, apperror.Forbidden("You do not have access to this task")
	}
	return acc, nil
}

func isAdminRole(role sqlc.NullRole) bool {
	if !role.Valid {
		return false
	}
	return role.Role == sqlc.RoleOWNER || role.Role == sqlc.RoleMANAGER || role.Role == sqlc.RoleSUPERADMIN
}

func safeInt32(n int64) int32 {
	if n < 0 {
		return 0
	}
	if n > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(n)
}

func categorizeMime(mime string) string {
	switch {
	case strings.HasPrefix(mime, "image/"):
		return "Images"
	case strings.Contains(mime, "pdf") || strings.Contains(mime, "document") || strings.Contains(mime, "word"):
		return "Documents"
	case strings.HasPrefix(mime, "text/"):
		return "Text Files"
	case strings.Contains(mime, "zip") || strings.Contains(mime, "rar") || strings.Contains(mime, "compressed"):
		return "Archives"
	default:
		return "Other"
	}
}

func (s *service) Upload(ctx context.Context, taskIdentifier string, userID uuid.UUID, fileName, mimeType string, data []byte) (*AttachmentResponse, error) {
	if len(data) > maxUploadBytes {
		return nil, apperror.BadRequest("File size cannot exceed 50MB")
	}
	if !allowedMimeTypes[mimeType] {
		return nil, apperror.BadRequest("File type not allowed")
	}
	acc, err := s.checkAccess(ctx, userID, taskIdentifier)
	if err != nil {
		return nil, err
	}

	folder := fmt.Sprintf("tasks/%s", acc.TaskID.String())
	saved, err := s.storage.SaveFile(folder, fileName, data)
	if err != nil {
		return nil, apperror.BadRequest("Failed to upload file: " + err.Error())
	}

	key := saved.Key
	newID, err := s.repo.Create(ctx, CreateParams{
		ID:         uuid.New(),
		FileName:   fileName,
		FileSize:   safeInt32(saved.Size),
		MimeType:   mimeType,
		URL:        saved.URL,
		StorageKey: &key,
		TaskID:     acc.TaskID,
		CreatedBy:  userID,
	})
	if err != nil {
		_ = s.storage.DeleteFile(key)
		return nil, apperror.Internal("Failed to persist attachment")
	}

	detail, err := s.repo.GetDetailByID(ctx, newID)
	if err != nil || detail == nil {
		return nil, apperror.Internal("Failed to fetch created attachment")
	}
	resp := mapDetailRow(detail, false)
	return &resp, nil
}

func (s *service) FindAll(ctx context.Context, taskIdentifier string, userID uuid.UUID) ([]AttachmentResponse, error) {
	acc, err := s.checkAccess(ctx, userID, taskIdentifier)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListByTaskID(ctx, acc.TaskID)
	if err != nil {
		return nil, apperror.Internal("Failed to list attachments")
	}
	out := make([]AttachmentResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapListRow(r, true))
	}
	return out, nil
}

func (s *service) GetTaskAttachments(ctx context.Context, taskIdentifier string, userID uuid.UUID) ([]AttachmentResponse, error) {
	acc, err := s.checkAccess(ctx, userID, taskIdentifier)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListByTaskID(ctx, acc.TaskID)
	if err != nil {
		return nil, apperror.Internal("Failed to list task attachments")
	}
	out := make([]AttachmentResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapListRow(r, false))
	}
	return out, nil
}

func (s *service) FindOne(ctx context.Context, id, userID uuid.UUID) (*AttachmentResponse, error) {
	detail, err := s.repo.GetDetailByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch attachment")
	}
	if detail == nil {
		return nil, apperror.NotFound("Task attachment not found")
	}
	if _, accErr := s.checkAccess(ctx, userID, detail.TaskID.String()); accErr != nil {
		return nil, accErr
	}
	resp := mapDetailRow(detail, true)
	return &resp, nil
}

func (s *service) StreamFile(ctx context.Context, id, userID uuid.UUID, w http.ResponseWriter, r *http.Request, isDownload bool) error {
	detail, err := s.repo.GetDetailByID(ctx, id)
	if err != nil {
		return apperror.Internal("Failed to fetch attachment")
	}
	if detail == nil {
		return apperror.NotFound("Attachment not found")
	}
	if _, accErr := s.checkAccess(ctx, userID, detail.TaskID.String()); accErr != nil {
		return accErr
	}

	if !isDownload && !previewableMimeTypes[detail.MimeType] {
		return apperror.BadRequest("File type not previewable")
	}

	key := detail.StorageKey.String
	if !detail.StorageKey.Valid || key == "" {
		if detail.Url.Valid && detail.Url.String != "" {
			key = strings.TrimPrefix(detail.Url.String, "/uploads/")
		}
	}
	if key == "" {
		return apperror.NotFound("Attachment not found")
	}

	return s.storage.StreamLocal(w, r, key, isDownload)
}

func (s *service) GetStats(ctx context.Context, taskIdentifier string, userID uuid.UUID) (*StatsResponse, error) {
	acc, err := s.checkAccess(ctx, userID, taskIdentifier)
	if err != nil {
		return nil, err
	}

	groups, err := s.repo.GroupStatsByMime(ctx, acc.TaskID)
	if err != nil {
		return nil, apperror.Internal("Failed to compute attachment stats")
	}
	recent, err := s.repo.CountRecent(ctx, acc.TaskID, time.Now().UTC().Add(-7*24*time.Hour))
	if err != nil {
		return nil, apperror.Internal("Failed to count recent attachments")
	}

	var totalFiles, totalBytes int64
	categories := make(map[string]CategoryStat)
	for _, g := range groups {
		totalFiles += g.CountFiles
		totalBytes += g.TotalSize
		cat := categorizeMime(g.MimeType)
		cur := categories[cat]
		cur.Count += g.CountFiles
		cur.TotalSize += g.TotalSize
		categories[cat] = cur
	}
	totalMB := math.Round((float64(totalBytes)/(1024*1024))*100) / 100

	return &StatsResponse{
		TotalAttachments: totalFiles,
		TotalSizeBytes:   totalBytes,
		TotalSizeMB:      totalMB,
		FileCategories:   categories,
		RecentUploads:    recent,
	}, nil
}

func (s *service) Remove(ctx context.Context, id, userID uuid.UUID) (*AttachmentResponse, error) {
	detail, err := s.repo.GetDetailByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch attachment")
	}
	if detail == nil {
		return nil, apperror.NotFound("Task attachment not found")
	}

	acc, err := s.checkAccess(ctx, userID, detail.TaskID.String())
	if err != nil {
		return nil, err
	}
	isAdmin := isAdminRole(acc.ProjectRole) || isAdminRole(acc.WorkspaceRole) || isAdminRole(acc.OrganizationRole)
	isOwner := detail.CreatedByID.Valid && uuid.UUID(detail.CreatedByID.Bytes) == userID
	if !isOwner && !isAdmin {
		return nil, apperror.Forbidden("You do not have permission to delete this attachment")
	}

	if detail.FilePath.Valid && detail.FilePath.String != "" {
		_ = os.Remove(detail.FilePath.String)
	}
	if detail.StorageKey.Valid && detail.StorageKey.String != "" {
		_ = s.storage.DeleteFile(detail.StorageKey.String)
	}

	if _, delErr := s.repo.DeleteByID(ctx, id); delErr != nil {
		return nil, apperror.Internal("Failed to delete attachment")
	}

	resp := mapDetailRow(detail, false)
	return &resp, nil
}
