package storage_files

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/platform/storage"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/httpx"
)

const defaultMaxAvatarBytes int64 = 50 * 1024 * 1024

// StorageBackend defines the storage operations required by HTTP upload/download handlers.
type StorageBackend interface {
	SaveFile(folder, originalName string, content []byte) (storage.SaveResult, error)
	Stream(w http.ResponseWriter, r *http.Request, key string, forceAttachment bool) error
	IsUsingS3() bool
}

// Handler serves public file downloads and avatar uploads under /api/uploads.
type Handler struct {
	store        StorageBackend
	maxAvatarSz  int64
	allowedTypes map[string]bool
}

// NewHandler constructs a storage files handler using config limits.
func NewHandler(store StorageBackend, cfg *config.Config) *Handler {
	maxSz := defaultMaxAvatarBytes
	typesList := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
	if cfg != nil {
		if cfg.MaxAvatarSize > 0 {
			maxSz = cfg.MaxAvatarSize
		}
		if len(cfg.AllowedImageTypes) > 0 {
			typesList = cfg.AllowedImageTypes
		}
	}

	allowed := make(map[string]bool)
	for _, t := range typesList {
		trimmed := strings.TrimSpace(strings.ToLower(t))
		if trimmed != "" {
			allowed[trimmed] = true
		}
	}

	return &Handler{
		store:        store,
		maxAvatarSz:  maxSz,
		allowedTypes: allowed,
	}
}

// ServeTaskFile handles GET /api/uploads/tasks/{taskId}/{filename}.
func (h *Handler) ServeTaskFile(w http.ResponseWriter, r *http.Request) {
	safeTaskID, err := sanitizePathComponent(chi.URLParam(r, "taskId"))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	safeName, err := sanitizePathComponent(chi.URLParam(r, "filename"))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if sErr := h.store.Stream(w, r, "tasks/"+safeTaskID+"/"+safeName, false); sErr != nil {
		httpx.RespondError(w, r, sErr)
	}
}

// ServeEditorImage handles GET /api/uploads/editor-images/{filename}.
func (h *Handler) ServeEditorImage(w http.ResponseWriter, r *http.Request) {
	safeName, err := sanitizePathComponent(chi.URLParam(r, "filename"))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if sErr := h.store.Stream(w, r, "editor-images/"+safeName, false); sErr != nil {
		httpx.RespondError(w, r, sErr)
	}
}

// ServeFile handles GET /api/uploads/{folder}/{filename}.
func (h *Handler) ServeFile(w http.ResponseWriter, r *http.Request) {
	safeFolder, err := sanitizePathComponent(chi.URLParam(r, "folder"))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	safeName, err := sanitizePathComponent(chi.URLParam(r, "filename"))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if sErr := h.store.Stream(w, r, safeFolder+"/"+safeName, false); sErr != nil {
		httpx.RespondError(w, r, sErr)
	}
}

// UploadFile handles POST /api/uploads/upload/{folder}, locked to "avatar".
func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	safeFolder, err := sanitizePathComponent(chi.URLParam(r, "folder"))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if safeFolder != "avatar" {
		httpx.RespondError(w, r, apperror.BadRequest("Unsupported upload folder: "+safeFolder))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxAvatarSz+(1<<20))
	if pErr := r.ParseMultipartForm(h.maxAvatarSz); pErr != nil {
		httpx.RespondError(w, r, apperror.BadRequest(fmt.Sprintf("File size exceeds %s limit", formatBytes(h.maxAvatarSz))))
		return
	}

	file, header, fErr := r.FormFile("file")
	if fErr != nil {
		httpx.RespondError(w, r, apperror.BadRequest("No file uploaded"))
		return
	}
	defer func() { _ = file.Close() }()

	content, readErr := io.ReadAll(io.LimitReader(file, h.maxAvatarSz+1))
	if readErr != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Failed reading uploaded file"))
		return
	}
	if int64(len(content)) > h.maxAvatarSz {
		httpx.RespondError(w, r, apperror.BadRequest(fmt.Sprintf("File size exceeds %s limit", formatBytes(h.maxAvatarSz))))
		return
	}

	mimeType := detectUploadMime(header.Header.Get("Content-Type"), content)
	if !h.allowedTypes[mimeType] {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid file type"))
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = extForMime(mimeType)
	}
	uniqueName := uuid.NewString() + ext

	saved, saveErr := h.store.SaveFile(safeFolder, uniqueName, content)
	if saveErr != nil {
		httpx.RespondError(w, r, saveErr)
		return
	}

	httpx.RespondJSON(w, http.StatusCreated, map[string]any{
		"message": "File uploaded successfully",
		"url":     saved.URL,
		"key":     saved.Key,
		"size":    saved.Size,
		"inCloud": h.store.IsUsingS3(),
	})
}

func sanitizePathComponent(component string) (string, error) {
	if component == "" {
		return "", apperror.BadRequest("Invalid path component")
	}
	cleaned := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(component, "..", ""), "/", ""), "\\", "")
	if cleaned != component {
		return "", apperror.BadRequest("Invalid characters in path")
	}
	return cleaned, nil
}

func detectUploadMime(headerMime string, content []byte) string {
	trimmed := strings.ToLower(strings.TrimSpace(strings.Split(headerMime, ";")[0]))
	if trimmed != "" && trimmed != "application/octet-stream" {
		return trimmed
	}
	return strings.ToLower(strings.TrimSpace(strings.Split(http.DetectContentType(content), ";")[0]))
}

func extForMime(m string) string {
	switch m {
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg"
	}
}

func formatBytes(bytes int64) string {
	const gb = 1024 * 1024 * 1024
	const mb = 1024 * 1024
	if bytes >= gb {
		return fmt.Sprintf("%.1fGB", float64(bytes)/float64(gb))
	}
	return fmt.Sprintf("%dMB", bytes/mb)
}
