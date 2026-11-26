package task_attachments

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles HTTP requests for task attachments.
type Handler struct {
	service Service
}

// NewHandler constructs a task_attachments HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

func extractUserID(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		return uuid.Nil, apperror.Unauthorized("Unauthenticated")
	}
	return id, nil
}

// Upload handles POST /api/task-attachments/upload/{taskId}.
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if parseErr := r.ParseMultipartForm(maxUploadBytes); parseErr != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid multipart payload or file too large"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("No file provided"))
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Failed to read uploaded file"))
		return
	}
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}
	taskID := chi.URLParam(r, "taskId")
	res, err := h.service.Upload(r.Context(), taskID, userID, header.Filename, mimeType, data)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// FindAll handles GET /api/task-attachments.
func (h *Handler) FindAll(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	taskID := httpx.QueryParam(r, "taskId", "")
	res, err := h.service.FindAll(r.Context(), taskID, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetStats handles GET /api/task-attachments/stats.
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	taskID := httpx.QueryParam(r, "taskId", "")
	res, err := h.service.GetStats(r.Context(), taskID, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetTaskAttachments handles GET /api/task-attachments/task/{taskId}.
func (h *Handler) GetTaskAttachments(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	taskID := chi.URLParam(r, "taskId")
	res, err := h.service.GetTaskAttachments(r.Context(), taskID, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindOne handles GET /api/task-attachments/{id}.
func (h *Handler) FindOne(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.FindOne(r.Context(), id, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Download handles GET /api/task-attachments/{id}/download.
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if err := h.service.StreamFile(r.Context(), id, userID, w, r, true); err != nil {
		httpx.RespondError(w, r, err)
	}
}

// Preview handles GET /api/task-attachments/{id}/preview.
func (h *Handler) Preview(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if err := h.service.StreamFile(r.Context(), id, userID, w, r, false); err != nil {
		httpx.RespondError(w, r, err)
	}
}

// Remove handles DELETE /api/task-attachments/{id}.
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if _, err := h.service.Remove(r.Context(), id, userID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
